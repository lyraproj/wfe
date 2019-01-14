package api

import "github.com/lyraproj/puppet-evaluator/eval"

type Resource interface {
	Activity

	Type() eval.ObjectType

	HandlerId() eval.TypedName

	ExtId() eval.Value
}
