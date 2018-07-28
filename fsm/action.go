package fsm

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-evaluator/utils"
	"github.com/puppetlabs/go-issues/issue"
	"gonum.org/v1/gonum/graph"
	"io"
	"reflect"
	"runtime"
)

type Action interface {
	issue.Named

	Call(Genesis, []reflect.Value) (map[string]reflect.Value, error)

	Consumes() []*Parameter

	Produces() []*Parameter
}

type ServerAction interface {
	Action
	graph.Node
	eval.PuppetObject

	SetResolved()

	// Resolved channel will be closed when the action is resolved
	Resolved() <-chan bool
}

var genesisType = reflect.TypeOf([]Genesis{}).Elem()
var errorType = reflect.TypeOf([]error{}).Elem()

func NewGoAction(c eval.Context, name string, function interface{}, paramNames []string) Action {
	ar := reflect.TypeOf(function)
	if ar.Kind() != reflect.Func {
		_, file, line, _ := runtime.Caller(1)
		panic(issue.NewReported(`GENESIS_ACTION_NOT_FUNCTION`, issue.SEVERITY_ERROR, issue.H{`name`: name, `type`: ar.String()}, issue.NewLocation(file, line, 0)))
	}

	paramc := len(paramNames)
	outc := ar.NumOut()
	if ar.IsVariadic() || !(ar.NumIn() - 1 == paramc && (outc == 1 && ar.Out(0).AssignableTo(errorType) || outc == 2 && ar.Out(1).AssignableTo(errorType) && ar.In(0).AssignableTo(genesisType))) {
		panic(badActionFunction(name, paramc, ar))
	}

	var produces []*Parameter
	if outc == 2 {
		retType := ar.Out(0)
		if retType.Kind() != reflect.Ptr {
			panic(badActionFunction(name, paramc, ar))
		}

		retType = retType.Elem()
		if retType.Kind() != reflect.Struct {
			panic(badActionFunction(name, paramc, ar))
		}

		outCount := retType.NumField()
		produces = make([]*Parameter, outCount)
		for i := 0; i < outCount; i++ {
			fld := retType.Field(i)
			produces[i] = NewParameter(name+`.`+utils.CamelToSnakeCase(fld.Name), eval.WrapType(c, fld.Type))
		}
	} else {
		produces = []*Parameter{}
	}

	consumes := make([]*Parameter, paramc)
	for i, n := range paramNames {
		consumes[i] = NewParameter(n, eval.WrapType(nil, ar.In(i+1)))
	}
	return NewAction(name, NewGoActionCall(function), consumes, produces)
}

func badActionFunction(name string, paramc int, typ reflect.Type) error {
	_, file, line, _ := runtime.Caller(2)
	return issue.NewReported(`GENESIS_ACTION_BAD_FUNCTION`, issue.SEVERITY_ERROR, issue.H{`name`: name, `param_count`: paramc, `type`: typ.String()}, issue.NewLocation(file, line, 0))
}

type Parameter struct {
	name string
	typ eval.PType
}

func (p *Parameter) Name() string {
	return p.name
}

func (p *Parameter) Type() eval.PType {
	return p.typ
}

func NewParameter(name string, typ eval.PType) *Parameter {
	return &Parameter{name, typ}
}

type BasicAction struct {
	name     string
	consumes []*Parameter
	produces []*Parameter
	function ActionFunction
}

type serverAction struct {
	Action
	graph.Node
	resolved chan bool
}

func NewAction(name string, function ActionFunction, consumes, produces []*Parameter) Action {
	return &BasicAction{name, consumes, produces, function}
}

func (a *BasicAction) Consumes() []*Parameter {
	return a.consumes
}

func (a *BasicAction) Produces() []*Parameter {
	return a.produces
}

func (a *BasicAction) Name() string {
	return a.name
}

func (a *BasicAction) Call(g Genesis, args []reflect.Value) (map[string]reflect.Value, error) {
	return a.function.Call(g, a, args)
}

func (a *serverAction) SetResolved() {
	close(a.resolved)
}

func (a *serverAction) Get(c eval.Context, key string) (value eval.PValue, ok bool) {
	switch key {
	case `name`:
		return types.WrapString(a.Name()), true
	case `id`:
		return types.WrapInteger(a.ID()), true
	default:
		return nil, false
	}
}

func (a *serverAction) Resolved() <-chan bool {
	return a.resolved
}

func (a *serverAction) InitHash() eval.KeyedValue {
	return types.SingletonHash2(`name`, types.WrapString(a.Name()))
}

var actionType eval.ObjectType

func init() {
	actionType = eval.NewObjectType(`Genesis::Action`, `{
		attributes => {
      name => String
    },
  }`)
}

func (a *serverAction) String() string {
	return eval.ToString(a)
}

func (a *serverAction) Equals(other interface{}, guard eval.Guard) bool {
	return a == other
}

func (a *serverAction) ToString(bld io.Writer, format eval.FormatContext, g eval.RDetect) {
	types.ObjectToString(a, format, bld, g)
}

func (a *serverAction) Type() eval.PType {
	return actionType
}

type ActionFunction interface {
	Call(g Genesis, a Action, args []reflect.Value) (map[string]reflect.Value, error)
}

type goActionCall struct {
	function interface{}
}

func NewGoActionCall(f interface{}) ActionFunction {
	return &goActionCall{f}
}

func (ga *goActionCall) Call(g Genesis, a Action, args []reflect.Value) (map[string]reflect.Value, error) {
	fv := reflect.ValueOf(ga.function)
	result := fv.Call(append([]reflect.Value{reflect.ValueOf(g)}, args...))
	expCount := 1
	if len(a.Produces()) > 1 {
		expCount++
	}
	rn := len(result)
	if rn != expCount {
		return nil, issue.NewReported(GENESIS_ACTION_BAD_RETURN_COUNT, issue.SEVERITY_ERROR, issue.H{`name`: a.Name(), `expected_count`: expCount, `actual_count`: rn}, nil)
	}

	if rn == 1 {
		if err := result[0].Interface(); err != nil {
			return nil, err.(error)
		}
		return nil, nil
	}

	err := result[1].Interface()
	if err != nil {
		return nil, err.(error)
	}

	rs := result[0]
	rt := rs.Type()
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rs = rs.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, issue.NewReported(GENESIS_ACTION_BAD_RETURN, issue.SEVERITY_ERROR, issue.H{`name`: a.Name(), `type`: rt.String()}, nil)
	}
	fcnt := rt.NumField()
	rmap := make(map[string]reflect.Value, fcnt)
	for i := 0; i < fcnt; i++ {
		ft := rt.Field(i)
		v := rs.Field(i)
		n := a.Name() + `.` + utils.CamelToSnakeCase(ft.Name)
		if v.IsValid() {
			rmap[n] = v
		} else {
			rmap[n] = reflect.Zero(ft.Type)
		}
	}
	return rmap, nil
}
