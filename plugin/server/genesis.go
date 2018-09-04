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
	stream fsmpb.Actors_InvokeActionServer
}

func NewGenesis(stream fsmpb.Actors_InvokeActionServer) api.Genesis {
	return &GRPCGenesis{Context: stream.Context(), stream: stream}
}

func (c *GRPCGenesis) call(id int64, args map[string]reflect.Value) map[string]reflect.Value {
	d, err := datapb.ToDataHash(args)
	if err != nil {
		panic(err)
	}

	if err := c.stream.Send(&fsmpb.Message{Id: id, Value: &datapb.Data{&datapb.Data_HashValue{d}}}); err != nil {
		panic(err)
	}

	resp, err := c.stream.Recv()
	if err != nil {
		// Even EOF is an error here
		panic(err)
	}

	if resp.Id != id {
		panic(fmt.Errorf("expected reply with id %d, got %d", id, resp.Id))
	}

	v, err := datapb.FromDataHash(resp.GetValue().GetHashValue())
	if err != nil {
		panic(err)
	}
	return v
}

func (c *GRPCGenesis) Resource(r map[string]reflect.Value) map[string]reflect.Value {
	return c.call(shared.GenesisResourceId, r)
}

func (c *GRPCGenesis) Notice(message string) {
	err := c.stream.Send(&fsmpb.Message{Id: shared.GenesisNoticeId, Value: &datapb.Data{Kind: &datapb.Data_StringValue{StringValue: message}}})
	if err != nil {
		panic(err)
	}
}

func (c *GRPCGenesis) ParentContext() context.Context {
	return c.Context
}
