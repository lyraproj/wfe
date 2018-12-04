package api

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/servicesdk/wfapi"
)

type Iterator interface {
	Activity

	// Style returns the style of iterator, times, range, each, or eachPair.
	IterationStyle() wfapi.IterationStyle

	// Producer returns the Activity that will be invoked once for each iteration
	Producer() Activity

	// Over returns what this iterator will iterate over. These parameters will be added
	// to the declared input set when the final requirements for the activity are computed.
	Over() []eval.Parameter

	// Variables returns the variables that this iterator will produce for each iteration. These
	// variables will be removed from the declared input set when the final requirements
	// for the activity are computed.
	Variables() []eval.Parameter
}
