package server

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"google.golang.org/grpc"
	"net/rpc"
	"github.com/puppetlabs/data-protobuf/datapb"
	"reflect"
)

type Genesis struct {
}

func (a *Genesis) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *Genesis) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for rpc`, a)
}

func (a *Genesis) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	return fmt.Errorf(`%T has no server implementation for grpc`, a)
}

func (a *Genesis) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCGenesis{ctx, fsmpb.NewGenesisClient(c)}, nil
}

type GRPCGenesis struct {
	context.Context
	client fsmpb.GenesisClient
}

func (c *GRPCGenesis) Apply(resources map[string]reflect.Value) map[string]reflect.Value {
	rh, err := datapb.ToDataHash(resources)
	if err != nil {
		panic(err)
	}
	resp, err := c.client.Apply(c, rh)
	if err != nil {
		panic(err)
	}
	vh, err := datapb.FromDataHash(resp)
	if err != nil {
		panic(err)
	}
	return vh
}
