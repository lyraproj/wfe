package wfe

import (
	"github.com/lyraproj/pcore/px"
)

type Resource interface {
	Step

	// Type returns the resource type
	Type() px.ObjectType

	// HandlerId returns the identifier of the handler of this type of resource
	HandlerId() px.TypedName

	// ExtId returns an explicitly defined external ID of the resource or nil if no
	// such id has been defined.
	ExtId() px.Value
}
