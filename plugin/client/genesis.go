package client

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"google.golang.org/grpc"
	"net/rpc"
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/plugin/shared"
)

type Genesis struct {
	context.Context
	impl shared.PbGenesis
}

func NewGenesis(ctx context.Context, impl shared.PbGenesis) *Genesis {
	return &Genesis{ctx, impl}
}

func (s *Genesis) Apply(resources *datapb.DataHash) *datapb.DataHash {
	return nil
}

func (a *Genesis) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *Genesis) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for rpc`, a)
}

func (a *Genesis) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	fsmpb.RegisterGenesisServer(s, &GRPCGenesis{impl: a})
	return nil
}

func (a *Genesis) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for grpc`, a)
}

type GRPCGenesis struct {
	impl *Genesis
}

func (s *GRPCGenesis) Apply(ctx context.Context, resources *datapb.DataHash) (*datapb.DataHash, error) {
	return nil, nil
}
