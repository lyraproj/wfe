package server

import (
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/fsm"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"github.com/puppetlabs/go-issues/issue"
	"golang.org/x/net/context"
	"reflect"
)

type grpcActionCall struct {
	client fsmpb.ActorClient
	id     int64
}

func NewGrpcActionFunction(client fsmpb.ActorClient, id int64) fsm.ActionFunction {
	return &grpcActionCall{client, id}
}

func (ga *grpcActionCall) Call(g fsm.Context, a fsm.Action, args map[string]reflect.Value) map[string]reflect.Value {
	nargs := len(args)
	if nargs != len(a.Consumes()) {
		panic(issue.NewReported(fsm.GENESIS_ACTION_BAD_CONSUMES_COUNT, issue.SEVERITY_ERROR, issue.H{`name`: a.Name(), `expected`: len(a.Consumes()), `actual`: nargs}, nil))
	}

	pe := make([]*datapb.DataEntry, nargs)
	for i, p := range a.Consumes() {
		dv, err := datapb.ToData(args[p.Name])
		if err != nil {
			panic(err)
		}
		pe[i] = &datapb.DataEntry{p.Name, dv}
	}
	result, err := ga.client.InvokeAction(g.(context.Context), &fsmpb.ActionInvocation{Id: ga.id, Arguments: &datapb.DataHash{Entries: pe}})
	if err != nil {
		panic(err)
	}

	re := make(map[string]reflect.Value, len(result.Entries))
	for _, r := range result.Entries {
		rv, err := datapb.FromData(r.Value)
		if err != nil {
			panic(err)
		}
		re[r.Key] = rv
	}
	return re
}
