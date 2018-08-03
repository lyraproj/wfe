package api

import (
	"reflect"
)

type Action interface {
	Name() string

	Consumes() []Parameter

	Produces() []Parameter

	Call(Genesis, map[string]reflect.Value) map[string]reflect.Value
}

type action struct {
	name     string
	consumes []Parameter
	produces []Parameter
	function ActionFunction
}

type ActionFunction interface {
	Call(Genesis, Action, map[string]reflect.Value) map[string]reflect.Value
}

func NewAction(name string, function ActionFunction, consumes, produces []Parameter) Action {
	return &action{name, consumes, produces, function}
}

func (a *action) Consumes() []Parameter {
	return a.consumes
}

func (a *action) Produces() []Parameter {
	return a.produces
}

func (a *action) Name() string {
	return a.name
}

func (a *action) Call(g Genesis, args map[string]reflect.Value) map[string]reflect.Value {
	return a.function.Call(g, a, args)
}
