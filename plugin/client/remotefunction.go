package client

import (
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/api"
	"reflect"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"github.com/puppetlabs/go-fsm/plugin/shared"
)

type remoteFunction struct {
	actors *GRPCActors
	actorName string
	actionName string
}

func NewRemoteAction(actors *GRPCActors, actorName string, action *fsmpb.Action) api.Action {
	return api.NewAction(
		action.Name,
		&remoteFunction{actors: actors, actorName: actorName, actionName: action.Name},
		shared.ConvertFromPbParams(action.Input),
		shared.ConvertFromPbParams(action.Output))
}

func (pf *remoteFunction) Call(g api.Genesis, a api.Action, args map[string]reflect.Value) map[string]reflect.Value {
	d, err := datapb.ToData(reflect.ValueOf([]interface{}{pf.actorName, pf.actionName, args}))
	if err != nil {
		panic(err)
	}
	d, err = pf.actors.InvokeAction(d, g)
	if err != nil {
		panic(err)
	}
	v, err := datapb.FromDataHash(d.GetHashValue())
	if err != nil {
		panic(err)
	}
	return v
}
