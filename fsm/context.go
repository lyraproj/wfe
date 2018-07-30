package fsm

import "context"

type Context interface {
	context.Context

	Action(name string, function interface{})
}
