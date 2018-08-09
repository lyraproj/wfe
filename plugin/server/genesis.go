package server

import (
	"context"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"github.com/puppetlabs/data-protobuf/datapb"
	"reflect"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/plugin/shared"
	"fmt"
)

type GRPCGenesis struct {
	context.Context
	stream fsmpb.Actor_InvokeActionServer
}

func NewGenesis(stream fsmpb.Actor_InvokeActionServer) api.Genesis {
	return &GRPCGenesis{stream.Context(), stream}
}

func (c *GRPCGenesis) Apply(resources map[string]reflect.Value) map[string]reflect.Value {
	rh, err := datapb.ToDataHash(resources)
	if err != nil {
		panic(err)
	}

	if err := c.stream.Send(&fsmpb.ActionMessage{shared.GenesisServiceId, rh}); err != nil {
		panic(err)
	}

	resp, err := c.stream.Recv()
	if err != nil {
		// Even EOF is an error here
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	if resp.Id != shared.GenesisServiceId {
		panic(fmt.Errorf("expected reply with id %d, got %d", shared.GenesisServiceId, resp.Id))
	}

	vh, err := datapb.FromDataHash(resp.GetArguments())
	if err != nil {
		panic(err)
	}
	return vh
}
