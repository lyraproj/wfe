package api

import (
	"reflect"
)

type Actor interface {
	GetActions() map[string]Action

	InvokeAction(name string, parameters map[string]reflect.Value, genesis Genesis) map[string]reflect.Value
}
