package fsm

import (
	"reflect"
	"context"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/hashicorp/go-hclog"
	"github.com/puppetlabs/data-protobuf/datapb"
)

type genesis struct {
	context.Context
}

func NewGenesis(ctx context.Context) api.Genesis {
	return &genesis{ctx}
}

func (g *genesis) Apply(resources map[string]reflect.Value) map[string]reflect.Value {
	return resources;
}

func (g *genesis) Lookup(keys []string) map[string]reflect.Value {
	result := make(map[string]reflect.Value, len(keys))
	for _, k := range keys {
		if v, ok := g.lookupOne(k); ok {
			result[k] = v;
		}
	}
	return result;
}

func (g *genesis) lookupOne(key string) (reflect.Value, bool) {
	if key == `test` {
		return reflect.ValueOf(`value of test`), true
	}
	return datapb.InvalidValue, false
}

func (g *genesis) Notice(message string) {
	hclog.Default().Info(message)
}
