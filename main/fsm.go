package main

import (
	"github.com/puppetlabs/go-fsm/fsm"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"github.com/puppetlabs/go-fsm/server"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"log"
	"io/ioutil"
	"os/exec"
	"os"
	"fmt"
	"reflect"
	"github.com/puppetlabs/data-protobuf/datapb"
	"net/rpc"
	"context"
)

const (
	address = "localhost:50051"
)

func mainGrpc() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := fsmpb.NewActorClient(conn)
	ctx := context.Background()
	actor, err := client.GetActions(ctx, &fsmpb.ActionsRequest{})
	if err != nil {
		log.Fatalf("could not get actor: %v", err)
	}

	g := server.NewContext(ctx)
	for _, action := range actor.Actions {
		g.AddAction(fsm.NewAction(action.Name, server.NewGrpcActionFunction(client, action.Id), action.Consumes, action.Produces))
	}
	g.Validate()
	g.Run()
}

type Actor interface {
	GetActions() []*fsmpb.Action
	InvokeAction(id int64, parameters *datapb.DataHash) *datapb.DataHash
}

type pluginFunction struct {
	actor Actor
	id int64
}

func (pf *pluginFunction) Call(g fsm.Context, a fsm.Action, args map[string]reflect.Value) map[string]reflect.Value {
	dh, err := datapb.ToDataHash(args)
	if err != nil {
		panic(err)
	}
	vh, err := datapb.FromDataHash(pf.actor.InvokeAction(pf.id, dh))
	if err != nil {
		panic(err)
	}
	return vh
}

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

type actor struct {

}

func (a *actor) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	panic("implement me")
}

func (a *actor) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	panic("RPC not currently supported")
}

func (a *actor) Server(*plugin.MuxBroker) (interface{}, error) {
	panic("RPC not currently supported")
}

// RPCClient is an implementation of Actor that talks over GRPC.
type GRPCClient struct{ client fsmpb.ActorClient }

func (c *GRPCClient) GetActions() []*fsmpb.Action {
	resp, err := c.client.GetActions(context.Background(), &fsmpb.ActionsRequest{})
	if err != nil {
		panic(err)
	}
	return resp.Actions
}

func (c *GRPCClient) InvokeAction(id int64, parameters *datapb.DataHash) *datapb.DataHash {
	resp, err := c.client.InvokeAction(context.Background(), &fsmpb.ActionInvocation{id, parameters})
	if err != nil {
		panic(err)
	}
	return resp
}

func (a *actor) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{fsmpb.NewActorClient(c)}, nil
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"actor": &actor{},
}

func main() {
	log.SetOutput(ioutil.Discard)

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins:         PluginMap,
		Cmd:             exec.Command("/home/thhal/tools/node-v10.7.0-linux-x64/bin/node",  "/home/thhal/git/genesis-js/src/genesis.js"),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})
	defer client.Kill()

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

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	actor := raw.(Actor)
	ctx := context.Background()
	actions := actor.GetActions()
	if err != nil {
		log.Fatalf("could not get actor: %v", err)
	}

	g := server.NewContext(ctx)
	for _, action := range actions {
		g.AddAction(fsm.NewAction(action.Name, &pluginFunction{actor, action.GetId()}, action.Consumes, action.Produces))
	}
	g.Validate()
	g.Run()
}
