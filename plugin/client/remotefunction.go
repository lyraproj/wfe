package client

import (
	"context"
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/plugin/shared"
	"reflect"
	"github.com/puppetlabs/go-fsm/fsmpb"
)

type remoteFunction struct {
	actor shared.PbActor
	id    int64
}

func NewRemoteAction(actor shared.PbActor, action *fsmpb.Action) api.Action {
	return api.NewAction(action.Name, &remoteFunction{actor, action.Id}, convertFromPbParams(action.Consumes), convertFromPbParams(action.Produces))
}

func (pf *remoteFunction) Call(g context.Context, a api.Action, args map[string]reflect.Value) map[string]reflect.Value {
	dh, err := datapb.ToDataHash(args)
	if err != nil {
		panic(err)
	}
	vh, err := datapb.FromDataHash(pf.actor.InvokeAction(pf.id, dh))
	if err != nil {
		panic(err)
	}
	return vh
}

func convertFromPbParams(params []*fsmpb.Parameter) []api.Parameter {
	ps := make([]api.Parameter, len(params))
	for i, p := range params {
		ps[i] = api.NewParameter(p.GetName(), p.GetType())
	}
	return ps
}
