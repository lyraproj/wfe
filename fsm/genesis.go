package fsm

import (
	"reflect"
	"context"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/hashicorp/go-hclog"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"io"
	"github.com/puppetlabs/go-evaluator/serialization"
	"github.com/puppetlabs/go-issues/issue"
)

type genesis struct {
	context.Context
}

func (g *genesis) String() string {
	return eval.ToString(g)
}

func (g *genesis) Equals(other interface{}, guard eval.Guard) bool {
	return g == other
}

func (g *genesis) ToString(bld io.Writer, format eval.FormatContext, rd eval.RDetect) {
	types.ObjectToString(g, format, bld, rd)
}

func (g *genesis) Type() eval.PType {
	return GenesisContext_Type
}

func (g *genesis) Get(c eval.Context, key string) (value eval.PValue, ok bool) {
	return nil, false
}

func (g *genesis) InitHash() eval.KeyedValue {
	return eval.EMPTY_MAP
}

var GenesisContext_Type eval.PType

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

func (g *genesis) ParentContext() context.Context {
	return g.Context
}

func (g *genesis) Call(c eval.Context, method string, args []eval.PValue, block eval.Lambda) (result eval.PValue, ok bool) {
	switch method {
	case `notice`:
		g.Notice(args[0].String())
		return eval.UNDEF, true
	case `resource`:
		return convertOutput(c, g.Resource(convertInput(c, args[0]))), true
	}
	return nil, false
}

func convertInput(c eval.Context, in eval.PValue) map[string]reflect.Value {
	inh := serialization.NewToDataConverter(c, eval.EMPTY_MAP).Convert(in)
	if eval.Equals(inh, eval.UNDEF) {
		return map[string]reflect.Value{}
	}
	if hash, ok := inh.(eval.KeyedValue); ok {
		rf := c.Reflector()
		hv := make(map[string]reflect.Value, hash.Len())
		hash.EachPair(func(k, v eval.PValue) {
			hv[k.String()] = rf.Reflect(v)
		})
		return hv
	}
	panic(c.Error(nil, api.GENESIS_NOT_STRING_HASH, issue.H{`type`: in.Type().String()}))
}

func convertOutput(c eval.Context, out map[string]reflect.Value) eval.PValue {
	entries := make([]*types.HashEntry, 0, len(out))
	for k, v := range out {
		entries = append(entries, types.WrapHashEntry2(k, eval.Wrap2(c, v)))
	}
	return serialization.NewFromDataConverter(c, eval.EMPTY_MAP).Convert(types.WrapHash(entries))
}

func init() {
	GenesisContext_Type = eval.NewObjectType(`Genesis::Context`, `{
    functions => {
      'notice' => Callable[ScalarData],
      'resource' => Callable[[Object], Object]
    }
  }`, func(ctx eval.Context, args []eval.PValue) eval.PValue {
	return NewGenesis(ctx).(eval.PValue)
})
}
