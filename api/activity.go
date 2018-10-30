package api

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-issues/issue"
)

// An Activity of a Workflow. The workflow is an Activity in itself and can be used in
// another Workflow.
type Activity interface {
	issue.Labeled

	// When returns an optional Condition that controls whether or not this activity participates
	// in the workflow.
	When() Condition

	// Identifier returns a string that uniquely identifies the activity within a resource. The string
	// is guaranteed to remain stable across invocations provided that no activity names, resource types
	// or iterator inputs changes within the parent chain of this Activity.
	Identifier() string

	// Style returns the activity style, 'workflow', 'resource', 'action', or the generic 'activity'.
	Style() string

	// Name returns the fully qualified name of the Activity
	Name() string

	// Input returns the input requirements for the Activity
	Input() []eval.Parameter

	// Output returns the definition of that this Activity will produce
	Output() []eval.Parameter
}

type RunnableActivity interface {
	Activity

	// Run will execute this Activity. The given input must match the declared Input. It will return
	// a value that corresponds to the Output declaration.
	//
	// The Scope of the given eval.Context must contain an ActivityContext named "genesis::context" prior to this call.
	Run(ctx eval.Context, input eval.OrderedMap) eval.OrderedMap
}
