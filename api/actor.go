package api

import (
	"reflect"
	"fmt"
)

type Actor interface {
	Action

	GoActorBuilder

	GetActions() map[string]Action

	InvokeAction(name string, parameters map[string]reflect.Value, genesis Genesis) map[string]reflect.Value
}

type BasicActor struct {
	action

	actions map[string]Action
}

func NewActor(name string, input, output interface{}) Actor {
	return &BasicActor{
		action: action{
			name: name,
			input: MakeParams(`input`, input),
			output: MakeParams(`output`, output),
		},
		actions: map[string]Action{},
	}
}

func (a *BasicActor) Action(name string, function interface{}) {
	a.actions[name] = NewGoAction(name, function)
}

func (a *BasicActor) Call(Genesis, map[string]reflect.Value) map[string]reflect.Value {
	panic("implement me")
}

func (a *BasicActor) GetActions() map[string]Action {
	return a.actions
}

func (a *BasicActor) InvokeAction(name string, parameters map[string]reflect.Value, genesis Genesis) map[string]reflect.Value {
	action, found := a.actions[name]
	if !found {
		panic(fmt.Errorf("no action with name '%s' in actor '%s'", name, a.name))
	}
	return action.Call(genesis, parameters)
}
