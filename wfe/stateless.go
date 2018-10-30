package wfe

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-issues/issue"
)

type stateless struct {
	Activity

	invocable eval.InvocableValue
}

func StatelessRef(c eval.Context, functionName string, w api.Condition) api.Activity {
	l, ok := eval.Load(c, eval.NewTypedName(eval.FUNCTION, functionName))
	if !ok {
		panic(eval.Error(eval.EVAL_UNKNOWN_FUNCTION, issue.H{`name`: functionName}))
	}
	return Stateless(c, l.(eval.Function), w)
}

func Stateless(c eval.Context, f eval.Function, w api.Condition) api.Activity {
	a := &stateless{invocable: f}
	a.Init(f.Name(), convertInput(f), convertOutput(f), w)
	return a
}

func Stateless2(name string, f eval.InvocableValue, input, output []eval.Parameter, w api.Condition) api.Activity {
	a := &stateless{invocable: f}
	a.Init(name, input, output, w)
	return a
}

func (s *stateless) Run(ctx eval.Context, input eval.OrderedMap) eval.OrderedMap {
	var err error
	val := eval.EMPTY_MAP
	switch GetOperation(ActivityContext(ctx)) {
	case api.Read:
		if len(s.Output()) > 0 {
			val, err = CallInvocable(ctx, s.invocable, s.Input(), input)
		}
	case api.Upsert:
		val, err = CallInvocable(ctx, s.invocable, s.Input(), input)
	}
	if err != nil {
		panic(err)
	}
	return val
}

func (s *stateless) Label() string {
	return ActivityLabel(s)
}

func (s *stateless) Style() string {
	return `stateless`
}

func getLambda(function eval.Function) eval.Lambda {
	ds := function.Dispatchers()
	if len(ds) != 1 {
		panic(eval.Error(WF_MULTIPLE_DISPATCHERS, issue.H{`name`: function.Name()}))
	}
	return ds[0]
}

func convertInput(function eval.Function) []eval.Parameter {
	return getLambda(function).Parameters()
}

func convertOutput(function eval.Function) []eval.Parameter {
	l := getLambda(function)
	s := l.Signature()
	rt := s.ReturnType()
	if rt == nil {
		return []eval.Parameter{}
	}
	if st, ok := rt.(*types.StructType); ok {
		es := st.Elements()
		ps := make([]eval.Parameter, len(es))
		for i, e := range es {
			ps[i] = impl.NewParameter(e.Name(), e.Type(), nil, false)
		}
		return ps
	}
	panic(eval.Error(WF_OUTPUT_NOT_STRUCT, issue.H{`type`: rt}))
}
