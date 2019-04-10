package wfe

import (
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/wfe/service"
)

type workflow struct {
	Activity

	activities []api.Activity
}

var DefinitionListType = types.NewArrayType(serviceapi.DefinitionMetaType, nil)

func (w *workflow) Run(ctx px.Context, input px.OrderedMap) px.OrderedMap {
	wf := NewWorkflowEngine(w)
	wf.Validate()
	return wf.Run(ctx, input)
}

func (w *workflow) Identifier() string {
	return ActivityId(w)
}

func (w *workflow) Label() string {
	return ActivityLabel(w)
}

func (w *workflow) Style() string {
	return `workflow`
}

func (w *workflow) WithIndex(index int) api.Activity {
	wc := workflow{}
	wc = *w // Copy by value
	wc.setIndex(index)
	return &wc
}

func Workflow(c px.Context, def serviceapi.Definition) api.Workflow {
	wf := &workflow{}
	wf.Init(def)
	activities := service.GetProperty(def, `activities`, DefinitionListType).(px.List)
	as := make([]api.Activity, activities.Len())
	activities.EachWithIndex(func(v px.Value, i int) { as[i] = CreateActivity(c, v.(serviceapi.Definition)) })
	wf.activities = as
	return wf
}

func (w *workflow) Activities() []api.Activity {
	return w.activities
}
