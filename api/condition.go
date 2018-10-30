package api

import (
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
)

// A Condition evaluates to true or false depending on its given input
type Condition interface {
	fmt.Stringer

	// Precedence returns the operator precedence for this Condition
	Precedence() int

	// IsTrue returns true if the given input satisfies the condition, false otherwise
	IsTrue(input eval.OrderedMap) bool

	// Returns all names in use by this condition and its nested conditions. The returned
	// slice is guaranteed to be unique and sorted alphabetically
	Names() []string
}
