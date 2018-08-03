package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-fsm/test/common"
	"github.com/puppetlabs/go-fsm/plugin/server"
	"github.com/puppetlabs/go-fsm/api"
)

type OutA struct {
	A string
	B int64
}

type InB struct {
	A string
	B int64
}

type OutB1 struct {
	C string
	D int64
}

type OutB2 struct {
	E string
	F int64
}

type InC struct {
	C string
	D int64
	E string
	F int64
}

func main() {
	actor := server.NewActor(context.Background())

	actor.Action("a", func(g api.Genesis) (*OutA, error) {
		return &OutA{`hello`, 4}, nil
	})

	actor.Action("b1", func(g api.Genesis, in *InB) (*OutB1, error) {
		return &OutB1{in.A + ` world`, in.B + 4}, nil
	})

	actor.Action("b2", func(g api.Genesis, in *InB) (*OutB2, error) {
		return &OutB2{in.A + ` earth`, in.B + 8}, nil
	})

	actor.Action("c", func(g api.Genesis, in *InC) error {
		fmt.Printf("%s, %d, %s, %d\n", in.C, in.D, in.E, in.F)
		return nil
	})

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: common.Handshake,
		Plugins: map[string]plugin.Plugin{
			"actor": actor,
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
