package activity

import (
	"github.com/puppetlabs/go-evaluator/eval"
)

type Definition interface {
	// Identity is the unique identity for this Definition
	Identity() Identity

	// ServiceId is the identifier of the service
	ServiceId() eval.TypedName

	// Name of the activity
	Name() string

	Properties() eval.OrderedMap
}

type Resource interface {
	Definition

	State() eval.OrderedMap
}

type Workflow interface {
	Definition

	Definitions() []Definition
}

type Action interface {
	Definition

	// Operations is a list of operations that are callable on this service
	Operations() []Operation
}
