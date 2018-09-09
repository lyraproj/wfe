package main

import (
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-fsm/test/common"
	"github.com/puppetlabs/go-fsm/lang/rpc/server"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
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

type OutC struct {
	R string
}

func main() {
	actor := api.NewWorkflow(`testing`, nil, &OutC{})

	actor.Action("a", func(g api.Genesis) (*OutA, error) {
		return &OutA{`hello`, 4}, nil
	}, nil)

	actor.Action("b1", func(g api.Genesis, in *InB) (*OutB1, error) {
		vs := g.Resource(eval.Wrap(map[string]interface{}{`a`: in.A + ` world`, `b`: in.B + 5}).(eval.KeyedValue))
		return &OutB1{vs.Get5(`a`, eval.UNDEF).String(), vs.Get5(`b`, eval.UNDEF).(*types.IntegerValue).Int()}, nil
	}, nil)

	actor.Action("b2", func(g api.Genesis, in *InB) (*OutB2, error) {
		return &OutB2{in.A + ` earth`, in.B + 8}, nil
	}, nil)

	actor.Action("c", func(g api.Genesis, in *InC) (*OutC, error) {
		return &OutC{fmt.Sprintf("%s, %d, %s, %d\n", in.C, in.D, in.E, in.F)}, nil
	}, nil)

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: common.Handshake,
		Plugins: map[string]plugin.Plugin{
			"actors": server.NewActorsPlugin(actor),
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
