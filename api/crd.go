package api

import "github.com/puppetlabs/go-evaluator/eval"

type ErrorConstant string

func (e ErrorConstant) Error() string {
	return string(e)
}

const NotFound = ErrorConstant(`not found`)

type CRD interface {
	// Create creates the managed subject.
	Create(ctx eval.Context, input eval.OrderedMap) (eval.OrderedMap, error)

	// Read reads the current state of the managed subject.
	Read(ctx eval.Context, input eval.OrderedMap) (eval.OrderedMap, error)

	// Delete deletes the managed subject.
	Delete(ctx eval.Context, input eval.OrderedMap) (eval.OrderedMap, error)
}

type CRUD interface {
	// Update updates the managed subject.
	Update(ctx eval.Context, input eval.OrderedMap) (eval.OrderedMap, error)
}
