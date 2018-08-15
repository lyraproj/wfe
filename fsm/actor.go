package fsm

import (
	"github.com/puppetlabs/go-fsm/api"
	"reflect"
	"fmt"
	"context"
)

type Actor struct {
	context.Context
	name string
	actions map[string]api.Action
}

func NewActor(ctx context.Context, name string) *Actor {
	return &Actor{ctx, name, make(map[string]api.Action, 7)}
}

func (a *Actor) Action(name string, function interface{}) {
	a.actions[name] = api.NewGoAction(name, function)
}

func (a *Actor) GetActions() map[string]api.Action {
	return a.actions
}

func (a *Actor) Name() string {
	return a.name
}

func (a *Actor) InvokeAction(actionName string, in map[string]reflect.Value, genesis api.Genesis) map[string]reflect.Value {
	action, found := a.actions[actionName]
	if !found {
		panic(fmt.Errorf("no action with name '%s' in actor '%s'", actionName, a.name))
	}

	return action.Call(genesis, in)
}

