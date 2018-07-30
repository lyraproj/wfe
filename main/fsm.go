package main

import (
	"github.com/puppetlabs/go-fsm/fsm"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"github.com/puppetlabs/go-fsm/server"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

const (
	address = "localhost:50051"
)

func main() {
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
