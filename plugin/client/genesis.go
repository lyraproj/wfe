package client

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"google.golang.org/grpc"
	"net/rpc"
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/api"
)

type Genesis struct {
	context.Context
	impl api.Genesis
}

func NewGenesis(ctx context.Context, impl api.Genesis) *Genesis {
	return &Genesis{ctx, impl}
}

func (g *Genesis) Apply(resources *datapb.DataHash) *datapb.DataHash {
	return nil
}

func (g *Genesis) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, g)
}

func (g *Genesis) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for rpc`, g)
}

func (g *Genesis) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	fsmpb.RegisterGenesisServer(s, &GRPCGenesis{impl: g})
	return nil
}

func (g *Genesis) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for grpc`, g)
}

type GRPCGenesis struct {
	impl *Genesis
}

func (s *GRPCGenesis) Apply(ctx context.Context, resources *datapb.DataHash) (*datapb.DataHash, error) {
	return nil, nil
}
