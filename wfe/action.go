package wfe

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/api"
)

type action struct {
	Activity

	crd api.CRD
}

func Action(name string, crd api.CRD, input, output []eval.Parameter, w api.Condition) api.Activity {
	a := &action{crd: crd}
	a.Init(name, input, output, w)
	return a
}

func (a *action) Run(c eval.Context, input eval.OrderedMap) eval.OrderedMap {
	ac := ActivityContext(c)
	op := GetOperation(ac)

	// TODO: Deal with identity mapping here when input declares externalId (mandatory for a provider)
	var err error

	val := eval.EMPTY_MAP
	switch op {
	case api.Create:
		val, err = a.crd.Create(c, input)
	case api.Upsert:
		val, err = a.crd.Read(c, input)
		if err == api.NotFound {
			val, err = a.crd.Create(c, input)
			break
		}
		input = input.Merge(val)
		fallthrough
	case api.Update:
		if crud, ok := a.crd.(api.CRUD); ok {
			val, err = crud.Update(c, input)
			break
		}
		val, err = a.crd.Delete(c, input)
		if err == nil {
			val, err = a.crd.Create(c, input)
		}
	case api.Delete:
		val, err = a.crd.Delete(c, input)
	default:
		if len(a.Output()) > 0 {
			val, err = a.crd.Read(c, input)
		}
	}
	if err != nil {
		panic(err)
	}
	return val
}

func GetOperation(ac eval.OrderedMap) api.Operation {
	if op, ok := ac.Get4(`operation`); ok {
		return api.NewOperation(op.String())
	}
	return api.Read
}

func CallInvocable(ctx eval.Context, invocable eval.InvocableValue, inputParams []eval.Parameter, input eval.OrderedMap) (val eval.OrderedMap, err error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(error); ok {
				err = e
			} else {
				panic(x)
			}
		}
	}()

	if cn, ok := invocable.(eval.CallNamed); ok {
		val = cn.CallNamed(ctx, nil, input).(eval.OrderedMap)
	} else {
		args := make([]eval.Value, len(inputParams))
		for i, p := range inputParams {
			arg := input.Get5(p.Name(), eval.UNDEF)
			if df, ok := arg.(types.Deferred); ok {
				arg = df.Resolve(ctx)
			}
			args[i] = arg
		}
		val = invocable.Call(ctx, nil, args...).(eval.OrderedMap)
	}
	return val, nil
}

func (a *action) Label() string {
	return ActivityLabel(a)
}

func (a *action) Style() string {
	return `action`
}
