package server

import (
	"fmt"
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/proto"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/lang/rpc/fsmpb"
	"github.com/puppetlabs/go-fsm/lang/rpc/shared"
)

type GRPCGenesis struct {
	ctx    eval.Context
	stream fsmpb.Actors_InvokeActionServer
}

func NewGenesis(ctx eval.Context, stream fsmpb.Actors_InvokeActionServer) api.Genesis {
	return &GRPCGenesis{ctx: ctx, stream: stream}
}

func (c *GRPCGenesis) call(id int64, args eval.OrderedMap) eval.OrderedMap {
	d := proto.ToPBData(args)
	if err := c.stream.Send(&fsmpb.Message{Id: id, Value: d}); err != nil {
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

	return proto.FromPBData(resp.GetValue()).(eval.OrderedMap)
}

func (c *GRPCGenesis) Resource(r eval.OrderedMap) eval.OrderedMap {
	return c.call(shared.GenesisResourceId, r)
}

func (c *GRPCGenesis) Notice(message string) {
	err := c.stream.Send(&fsmpb.Message{Id: shared.GenesisNoticeId, Value: &datapb.Data{Kind: &datapb.Data_StringValue{StringValue: message}}})
	if err != nil {
		panic(err)
	}
}

func (c *GRPCGenesis) Context() eval.Context {
	return c.ctx
}
