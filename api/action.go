package api

import (
	"reflect"
)

type Action interface {
	Name() string

	Input() []Parameter

	Output() []Parameter

	Call(Genesis, map[string]reflect.Value) map[string]reflect.Value
}

type action struct {
	name     string
	input []Parameter
	output []Parameter
	function ActionFunction
}

type ActionFunction interface {
	Call(Genesis, Action, map[string]reflect.Value) map[string]reflect.Value
}

func NewAction(name string, function ActionFunction, input, output []Parameter) Action {
	return &action{name, input, output, function}
}

func (a *action) Input() []Parameter {
	return a.input
}

func (a *action) Output() []Parameter {
	return a.output
}

func (a *action) Name() string {
	return a.name
}

func (a *action) Call(g Genesis, args map[string]reflect.Value) map[string]reflect.Value {
	return a.function.Call(g, a, args)
}
