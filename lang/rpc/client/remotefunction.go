package client

import (
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/lang/rpc/fsmpb"
	"github.com/puppetlabs/go-fsm/lang/rpc/shared"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-evaluator/proto"
)

type remoteFunction struct {
	actors *GRPCActors
	actorName string
	actionName string
}

func NewRemoteAction(actors *GRPCActors, actorName string, action *fsmpb.Action) api.Activity {
	return api.NewAction(
		action.Name,
		&remoteFunction{actors: actors, actorName: actorName, actionName: action.Name},
		shared.ConvertIterate(action.Iterate),
		shared.ConvertFromPbParams(action.Input),
		shared.ConvertFromPbParams(action.Output))
}

func (pf *remoteFunction) Call(g api.Genesis, a api.Activity, args eval.KeyedValue) eval.PValue {
	els := []eval.PValue{types.WrapString(pf.actorName), types.WrapString(pf.actionName), args}
	d := proto.ToPBData(types.WrapArray(els))
	d, err := pf.actors.InvokeAction(d, g)
	if err != nil {
		panic(err)
	}
	return proto.FromPBData(d)
}
