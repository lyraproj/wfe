package puppet

import (
	"github.com/puppetlabs/go-parser/parser"
	"github.com/puppetlabs/go-evaluator/eval"
	"io"
	"github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-fsm/api"
	"reflect"
	"github.com/puppetlabs/go-issues/issue"
	"runtime"
)

type actor struct {
	basicAction
	actions map[string]api.Action
}

func (a *actor) Call(api.Genesis, map[string]reflect.Value) map[string]reflect.Value {
	panic("implement me")
}

func (a *actor) Action(name string, function interface{}) {
	panic("implement me")
}

func (a *actor) GetActions() map[string]api.Action {
	return a.actions
}

func (a *actor) InvokeAction(name string, parameters map[string]reflect.Value, genesis api.Genesis) map[string]reflect.Value {
	if action, ok := a.actions[name]; ok {
		return action.Call(genesis, parameters)
	}
	_, file, line, _ := runtime.Caller(1)
	panic(issue.NewReported(api.GENESIS_NO_SUCH_ACTION, issue.SEVERITY_ERROR, issue.H{`actor`: a.Name(), `action`: name}, issue.NewLocation(file, line, 0)))
}

func init() {
	impl.NewPuppetActor = NewPuppetActor
}

func NewPuppetActor(expr *parser.PlanDefinition) interface{} {
	a := &actor{}
	a.Init(&expr.FunctionDefinition)
	return a
}

func (a *actor) ToString(bld io.Writer, format eval.FormatContext, g eval.RDetect) {
	io.WriteString(bld, `actor `)
	io.WriteString(bld, a.Name())
}

func (a *actor) String() string {
	return eval.ToString(a)
}

func (a *actor) Resolve(c eval.Context) {
	a.PuppetFunction.Resolve(c)

	// Ensure that body consists of functions only and that all functions
	// are valid actions (or actors)
	body := a.Expression().Body().(*parser.BlockExpression).Statements()
	actions := make(map[string]api.Action, len(body))

	for _, s := range body {
		switch s.(type) {
		case *parser.PlanDefinition:
			pd := s.(*parser.PlanDefinition)
			if pd.Actor() {
				a := NewPuppetActor(pd)
				a.(eval.Resolvable).Resolve(c)
				actions[pd.Name()] = a.(api.Actor)
				continue
			}
		case *parser.FunctionDefinition:
			fd := s.(*parser.FunctionDefinition)
			a := NewAction(fd)
			a.(eval.Resolvable).Resolve(c)
			actions[a.Name()] = a
			continue
		}
		panic(`puppet actor contains non-function expressions`)
	}
	a.actions = actions
	a.input = a.convertInput(c, false)
	a.output = a.convertOutput(c)
}
