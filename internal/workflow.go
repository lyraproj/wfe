package internal

import (
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/wfe"
)

type workflow struct {
	step

	steps []wfe.Step
}

var DefinitionListType = types.NewArrayType(serviceapi.DefinitionMetaType, nil)

func (w *workflow) Run(ctx px.Context, parameters px.OrderedMap) px.OrderedMap {
	wf := newWorkflowEngine(w)
	wf.Validate()
	return wf.Run(ctx, parameters)
}

func (w *workflow) Identifier() string {
	return StepId(w)
}

func (w *workflow) Label() string {
	return StepLabel(w)
}

func (w *workflow) Style() string {
	return `workflow`
}

func (w *workflow) WithIndex(index int) wfe.Step {
	wc := *w // Copy by value
	wc.setIndex(index)
	return &wc
}

func newWorkflow(c px.Context, def serviceapi.Definition) wfe.Workflow {
	wf := &workflow{}
	wf.initStep(def)
	steps := wfe.GetProperty(def, `steps`, DefinitionListType).(px.List)
	as := make([]wfe.Step, steps.Len())
	steps.EachWithIndex(func(v px.Value, i int) { as[i] = CreateStep(c, v.(serviceapi.Definition)) })
	wf.steps = as
	return wf
}

func (w *workflow) Steps() []wfe.Step {
	return w.steps
}
