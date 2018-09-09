package golang

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/wfe"
)

func Iterator(ctx eval.Context, style api.IterationStyle, name string, function interface{}, over, variables interface{}) api.Activity {
  a := Activity(ctx, name, function)
	op := MakeParams(ctx, name, over)
	vp := MakeParams(ctx, name, variables)
	return wfe.Iterator(style, a, name, op, vp)
}

func Range(ctx eval.Context, name string, function interface{}, over, variables interface{}) api.Activity {
	return Iterator(ctx, api.IterationStyleRange, name, function, over, variables)
}

func Times(ctx eval.Context, name string, function interface{}, over, variables interface{}) api.Activity {
	return Iterator(ctx, api.IterationStyleTimes, name, function, over, variables)
}
