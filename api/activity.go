package api

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/wf"
)

// An Activity of a Workflow. The workflow is an Activity in itself and can be used in
// another Workflow.
type Activity interface {
	issue.Labeled

	// When returns an optional Condition that controls whether or not this activity participates
	// in the workflow.
	When() wf.Condition

	// Identifier returns a string that uniquely identifies the activity within a resource. The string
	// is guaranteed to remain stable across invocations provided that no activity names, resource types
	// or iterator inputs changes within the parent chain of this Activity.
	Identifier() string

	// The Id of the service that provides this activity
	ServiceId() px.TypedName

	// Style returns the activity style, 'workflow', 'resource', 'stateHandler', or 'action'.
	Style() string

	// Name returns the fully qualified name of the Activity
	Name() string

	// Input returns the input requirements for the Activity
	Input() []px.Parameter

	// Output returns the definition of that this Activity will produce
	Output() []px.Parameter

	// Run will execute this Activity. The given input must match the declared Input. It will return
	// a value that corresponds to the Output declaration.
	Run(ctx px.Context, input px.OrderedMap) px.OrderedMap
}
