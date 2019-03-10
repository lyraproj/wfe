package api

import "github.com/lyraproj/pcore/px"

type Resource interface {
	Activity

	Type() px.ObjectType

	HandlerId() px.TypedName

	ExtId() px.Value
}
