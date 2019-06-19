package wfe

type Workflow interface {
	Step

	// Steps returns the steps that constitutes this workflow.
	Steps() []Step
}
