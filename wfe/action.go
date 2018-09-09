package wfe

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-issues/issue"
)

type action struct {
	Activity

	invocable eval.InvocableValue
}

func ActionRef(c eval.Context, functionName string, w api.Condition) api.Activity {
	l, ok := eval.Load(c, eval.NewTypedName(eval.FUNCTION, functionName))
	if !ok {
		panic(eval.Error(eval.EVAL_UNKNOWN_FUNCTION, issue.H{`name`: functionName}))
	}
	return Action(c, l.(eval.Function), w)
}

func Action(c eval.Context, f eval.Function, w api.Condition) api.Activity {
	a := &action{invocable: f}
	a.Init(f.Name(), convertInput(f), convertOutput(f), w)
	return a
}

func Action2(name string, f eval.InvocableValue, input, output []eval.Parameter, w api.Condition) api.Activity {
	a := &action{invocable: f}
	a.Init(name, input, output, w)
	return a
}

func (a *action) Run(ctx eval.Context, input eval.KeyedValue) eval.KeyedValue {
	if cn, ok := a.invocable.(eval.CallNamed); ok {
		return cn.CallNamed(ctx, nil, input).(eval.KeyedValue)
	}
	s := a.Input()
	args := make([]eval.PValue, len(s))
	for i, p := range s {
		arg := input.Get5(p.Name(), eval.UNDEF)
		if df, ok := arg.(types.Deferred); ok {
			arg = df.Resolve(ctx)
		}
		args[i] = arg
	}
	return a.invocable.Call(ctx, nil, args...).(eval.KeyedValue)
}

func (a *action) Label() string {
	return ActivityLabel(a)
}

func (a *action) Style() string {
	return `action`
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
