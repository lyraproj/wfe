package wfe

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-servicesdk/service"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
)

type workflow struct {
	Activity

	activities []api.Activity
}

var DefinitionListType = types.NewArrayType(service.Definition_Type, nil)

func (w *workflow) Run(ctx eval.Context, input eval.OrderedMap) eval.OrderedMap {
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
	activities := GetProperty(def, `activities`, DefinitionListType).(eval.List)
	as := make([]api.Activity, activities.Len())
	activities.EachWithIndex(func(v eval.Value, i int) { as[i] = CreateActivity(v.(serviceapi.Definition))})
	w.activities = as
}
