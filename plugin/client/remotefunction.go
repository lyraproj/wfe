package client

import (
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/api"
	"reflect"
	"github.com/puppetlabs/go-fsm/fsmpb"
)

type remoteFunction struct {
	actor *GRPCActor
	id    int64
}

func NewRemoteAction(actor *GRPCActor, action *fsmpb.Action) api.Action {
	return api.NewAction(action.Name, &remoteFunction{actor, action.Id}, convertFromPbParams(action.Input), convertFromPbParams(action.Output))
}

func (pf *remoteFunction) Call(g api.Genesis, a api.Action, args map[string]reflect.Value) map[string]reflect.Value {
	dh, err := datapb.ToDataHash(args)
	if err != nil {
		panic(err)
	}
	dh, err = pf.actor.InvokeAction(pf.id, dh, g)
	if err != nil {
		panic(err)
	}
	vh, err := datapb.FromDataHash(dh)
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
