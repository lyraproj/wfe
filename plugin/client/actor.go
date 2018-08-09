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
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/plugin/shared"
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

	actor := raw.(*GRPCActor)
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

func (c *GRPCActor) InvokeAction(id int64, parameters *datapb.DataHash, genesis api.Genesis) (*datapb.DataHash, error) {
	stream, err := c.client.InvokeAction(c.ctx)
	if err != nil {
		return nil, err
	}

	err = stream.Send(&fsmpb.ActionMessage{Id: id, Arguments: parameters})
	for {
		resp, err := stream.Recv()
		if err != nil {
			// Even EOF is an error here
			return nil, err
		}
		switch resp.Id {
		case id:
			// This is the response for the InvokeAction call
			stream.CloseSend()
			return resp.Arguments, nil

		case shared.GenesisServiceId:
			// Message intended for the Genesis service
			oh, err := datapb.FromDataHash(resp.Arguments)
			if err != nil {
				return nil, err
			}
			dh, err := datapb.ToDataHash(genesis.Apply(oh))
			if err != nil {
				return nil, err
			}
			stream.Send(&fsmpb.ActionMessage{shared.GenesisServiceId, dh})

		default:
			return nil, fmt.Errorf("unexpected response id in ActionMessage: %d", resp.Id)
		}
	}
}
