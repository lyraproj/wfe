package api

import (
	"reflect"
	"context"
)

type Genesis interface {
	context.Context

	Apply(resources map[string]reflect.Value) map[string]reflect.Value
}
