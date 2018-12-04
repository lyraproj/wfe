package wfe

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wfapi"
)

type action struct {
	Activity
	typ eval.ObjectType
}

func Action(def serviceapi.Definition) api.Activity {
	a := &action{}
	a.Init(def)
	return a
}

func (a *action) Init(def serviceapi.Definition) {
	a.Activity.Init(def)
	// TODO: Type validation. The typ must be an ObjectType implementing read, upsert, and delete.
	a.typ = GetProperty(def, `interface`, types.NewTypeType(types.DefaultObjectType())).(eval.ObjectType)
}

func (a *action) Run(c eval.Context, input eval.OrderedMap) eval.OrderedMap {
	ac := ActivityContext(c)
	op := GetOperation(ac)
	invokable := a.GetService(c)

	switch op {
	case wfapi.Read:
		return invokable.Invoke(c, a.name, `read`, input).(eval.OrderedMap)

	case wfapi.Upsert:
		return invokable.Invoke(c, a.name, `upsert`, input).(eval.OrderedMap)

	case wfapi.Delete:
		invokable.Invoke(c, a.name, `delete`, input)
		return eval.EMPTY_MAP
	default:
		panic(eval.Error(wfapi.WF_ILLEGAL_OPERATION, issue.H{`operation`: op}))
	}
}

func GetOperation(ac eval.OrderedMap) wfapi.Operation {
	if op, ok := ac.Get4(`operation`); ok {
		return wfapi.Operation(op.(*types.IntegerValue).Int())
	}
	return wfapi.Read
}

func (a *action) Label() string {
	return ActivityLabel(a)
}

func (a *action) Style() string {
	return `action`
}
