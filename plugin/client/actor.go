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
	"reflect"
)

type ActorsPlugin struct {
}

func (a *ActorsPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *ActorsPlugin) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for rpc`, a)
}

func (a *ActorsPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	return fmt.Errorf(`%T has no server implementation for grpc`, a)
}

func (a *ActorsPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCActors{ctx, broker, fsmpb.NewActorsClient(c)}, nil
}

func RunActor(actorName string, client *plugin.Client, input map[string]reflect.Value) map[string]reflect.Value {
	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("actors")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	actors := raw.(*GRPCActors)
	ctx := context.Background()
	actor := actors.GetActor(actorName)
	if err != nil {
		log.Fatalf("could not get actor: %v", err)
	}

	g := fsm.NewActorServer(ctx, actorName, shared.ConvertFromPbParams(actor.Input), shared.ConvertFromPbParams(actor.Output))
	for _, action := range actor.Actions {
		g.AddAction(NewRemoteAction(actors, actorName, action))
	}
	g.Validate()
	return g.Call(nil, input)
}

type GRPCActors struct {
	ctx    context.Context
	broker *plugin.GRPCBroker
	client fsmpb.ActorsClient
}

func (c *GRPCActors) GetActor(name string) *fsmpb.Actor {
	resp, err := c.client.GetActor(c.ctx, &fsmpb.ActorRequest{Name: name})
	if err != nil {
		panic(err)
	}
	return resp
}

func (c *GRPCActors) InvokeAction(args *datapb.Data, genesis api.Genesis) (*datapb.Data, error) {
	stream, err := c.client.InvokeAction(c.ctx)
	if err != nil {
		return nil, err
	}

	err = stream.Send(&fsmpb.Message{Id: shared.InvokeActionId, Value: args})
	for {
		resp, err := stream.Recv()
		if err != nil {
			// Even EOF is an error here
			return nil, err
		}
		switch resp.Id {
		case shared.InvokeActionId:
			// This is the response for the InvokeAction call
			stream.CloseSend()
			return resp.GetValue(), nil

		case shared.GenesisApplyId:
			// Message intended for the Genesis service
			v, err := datapb.FromDataHash(resp.GetValue().GetHashValue())
			if err != nil {
				return nil, err
			}
			d, err := datapb.ToData(reflect.ValueOf(genesis.Apply(v)))
			if err != nil {
				return nil, err
			}
			stream.Send(&fsmpb.Message{Id: resp.Id, Value: d})

		case shared.GenesisNoticeId:
			v, err := datapb.FromData(resp.GetValue())
			if err != nil {
				return nil, err
			}
			genesis.Notice(v.String())

		default:
			return nil, fmt.Errorf("unexpected response id in ActionMessage: %d", resp.Id)
		}
	}
}
