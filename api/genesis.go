package api

import (
	"reflect"
	"context"
)

type Genesis interface {
	context.Context

	Resource(map[string]reflect.Value) map[string]reflect.Value

	Notice(message string)
}
