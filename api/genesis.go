package api

import "reflect"

type Genesis interface {
	Apply(resources map[string]reflect.Value) map[string]reflect.Value
}
