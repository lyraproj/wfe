package fsm

import (
	"reflect"
	"context"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/hashicorp/go-hclog"
)

type genesis struct {
	context.Context
}

func NewGenesis(ctx context.Context) api.Genesis {
	return &genesis{ctx}
}

func (g *genesis) Resource(r map[string]reflect.Value) map[string]reflect.Value {
	// TODO: Really apply resource
	return r
}

func (g *genesis) Notice(message string) {
	hclog.Default().Info(message)
}
