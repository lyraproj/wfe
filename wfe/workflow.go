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

func (w *workflow) Label() string {
	return ActivityLabel(w)
}

func (w *workflow) Style() string {
	return `workflow`
}

func Workflow(def serviceapi.Definition) api.Workflow {
	wf := &workflow{}
	wf.Init(def)
	return wf
}

func (w *workflow) Activities() []api.Activity {
	return w.activities
}

func (w *workflow) Init(def serviceapi.Definition) {
	w.Activity.Init(def)
	activities := service.GetProperty(def, `activities`, DefinitionListType).(px.List)
	as := make([]api.Activity, activities.Len())
	activities.EachWithIndex(func(v px.Value, i int) { as[i] = CreateActivity(v.(serviceapi.Definition)) })
	w.activities = as
}
