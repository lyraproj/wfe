package wfe

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-fsm/api"
)

type workflow struct {
	Activity

	activities []api.Activity
}

func (w *workflow) Run(ctx eval.Context, input eval.KeyedValue) eval.KeyedValue {
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

func NewWorkflow(name string, input, output []eval.Parameter, w api.Condition, activities ...api.Activity) api.Workflow {
	wf := &workflow{activities: make([]api.Activity, len(activities))}
	copy(wf.activities, activities)
	wf.Init(name, input, output, w)
	return wf
}

func (w *workflow) Activities() []api.Activity {
	return w.activities
}
