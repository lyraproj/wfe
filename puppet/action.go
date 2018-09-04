package puppet

import (
	"reflect"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-evaluator/types"
	"runtime"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-parser/parser"
	"github.com/puppetlabs/go-fsm/fsm"
)

type basicAction struct {
	impl.PuppetFunction
	input []api.Parameter
	output []api.Parameter
}

func (a *basicAction) Input() []api.Parameter {
	return a.input
}

func (a *basicAction) Output() []api.Parameter {
	return a.output
}

func (a *basicAction) convertInput(c eval.Context, firstIsCtx bool) []api.Parameter {
	s := a.Parameters()
	np := len(s)
	if firstIsCtx {
		np--
	}
	ps := make([]api.Parameter, 0, np)
	for i, p := range s {
		if i == 0 && firstIsCtx {
			// TODO: Check that first parameter is of correct type
			continue
		}
		de := p.Default()
		var lookupArg *reflect.Value
		if de != nil {
			switch de.(type) {
			case *parser.CallNamedFunctionExpression:
				cf := de.(*parser.CallNamedFunctionExpression)
				if qn, ok := cf.Functor().(*parser.QualifiedName); ok && qn.Name() == `lookup` && len(cf.Arguments()) == 1 {
					// The actual lookup call is not evaluated here but the argument is
					arg := c.Reflector().Reflect(c.Evaluate(cf.Arguments()[0]))
					lookupArg = &arg
				}
			}
		}
		ps = append(ps, api.NewParameter(p.Name(), p.Type().String(), lookupArg))
	}
	return ps
}

func (a *basicAction) convertOutput(c eval.Context) []api.Parameter {
	rt := a.Signature().ReturnType()
	if rt == nil {
		return []api.Parameter{}
	}
	if st, ok := rt.(*types.StructType); ok {
		es := st.Elements()
		ps := make([]api.Parameter, len(es))
		for i, e := range es {
			ps[i] = api.NewParameter(e.Name(), e.Type().String(), nil)
		}
		return ps
	}
	panic(c.Error(a.Expression().ReturnType(), api.GENESIS_ACTION_NOT_STRUCT, issue.H{`name`: a.Name()}))
}

type producerAction struct {
	basicAction
	producer *actionCall
}

func NewAction(expr *parser.FunctionDefinition) api.Action {
	a := &producerAction{}
	a.Init(expr)
	return a
}

func (a *producerAction) Call(g api.Genesis, in map[string]reflect.Value) map[string]reflect.Value {
	return a.producer.Call(g, a, in)
}

func (a *producerAction) Resolve(c eval.Context) {
	a.PuppetFunction.Resolve(c)
	a.producer = &actionCall{&a.PuppetFunction}
	a.input = a.convertInput(c, true)
	a.output = a.convertOutput(c)
}

type actionCall struct {
	f *impl.PuppetFunction
}

func (pa *actionCall) Call(g api.Genesis, a api.Action, in map[string]reflect.Value) map[string]reflect.Value {
	c, ok := g.ParentContext().(eval.Context)
	if !ok {
		_, file, line, _ := runtime.Caller(1)
		panic(issue.NewReported(api.GENESIS_NOT_PUPPET_CONTEXT, issue.SEVERITY_ERROR, issue.NO_ARGS, issue.NewLocation(file, line, 0)))
	}
	c = c.Fork()

	params := make([]eval.PValue, len(a.Input()) + 1)
	params[0] = fsm.NewGenesis(g).(eval.PValue)
	for i, p := range a.Input() {
		params[i+1] = eval.Wrap2(c, in[p.Name()])
	}
	result := pa.f.Call(c, nil, params...)
	hash, ok := result.(*types.HashValue)
	if !ok {
		panic(eval.Error(c, api.GENESIS_NOT_STRING_HASH, issue.H{`type`: result.Type().String()}))
	}

	r := c.Reflector()
	vm := make(map[string]reflect.Value, hash.Len())
	hash.EachPair(func(k, v eval.PValue) {
		var ks *types.StringValue
		ks, ok = k.(*types.StringValue)
		if !ok {
			panic(eval.Error(c, api.GENESIS_NOT_STRING_HASH, issue.H{`type`: result.Type().String()}))
		}
		vm[ks.String()] = r.Reflect(v)
	})
	return vm;
}
