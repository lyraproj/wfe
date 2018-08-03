package client

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-fsm/fsm"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"google.golang.org/grpc"
	"log"
	"net/rpc"
	"os"
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/plugin/shared"
	"github.com/puppetlabs/go-fsm/api"
)

type Actor struct {
}

func (a *Actor) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *Actor) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for rpc`, a)
}

func (a *Actor) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	return fmt.Errorf(`%T has no server implementation for grpc`, a)
}

func (a *Actor) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCActor{ctx, broker, fsmpb.NewActorClient(c)}, nil
}

func RunActions(client *plugin.Client) {
	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("actor")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	actor := raw.(shared.PbActor)
	ctx := context.Background()
	actions := actor.GetActions()
	if err != nil {
		log.Fatalf("could not get actor: %v", err)
	}

	g := fsm.NewActorServer(ctx)
	for _, action := range actions {
		g.AddAction(NewRemoteAction(actor, action))
	}
	g.Validate()
	g.Run()
	g.DumpVariables()
}

type GRPCActor struct {
	ctx    context.Context
	broker *plugin.GRPCBroker
	client fsmpb.ActorClient
}

func (c *GRPCActor) GetActions() []*fsmpb.Action {
	resp, err := c.client.GetActions(c.ctx, &fsmpb.ActionsRequest{})
	if err != nil {
		panic(err)
	}
	return resp.Actions
}

func (c *GRPCActor) InvokeAction(id int64, parameters *datapb.DataHash, genesis api.Genesis) *datapb.DataHash {
	genesisServer := &GRPCGenesis{impl: NewGenesis(c.ctx, genesis)}
	var s *grpc.Server
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s = grpc.NewServer(opts...)
		fsmpb.RegisterGenesisServer(s, genesisServer)
		return s
	}

	brokerID := c.broker.NextId()
	go c.broker.AcceptAndServe(brokerID, serverFunc)

	resp, err := c.client.InvokeAction(c.ctx, &fsmpb.ActionInvocation{Genesis: brokerID, Id: id, Arguments: parameters})
	if err != nil {
		panic(err)
	}

	s.Stop()
	return resp
}
