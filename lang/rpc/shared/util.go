package shared

import (
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/proto"
	"github.com/puppetlabs/go-fsm/lang/rpc/fsmpb"
	"reflect"
	"github.com/puppetlabs/go-evaluator/impl"
)

func ExpandStringMap(sr reflect.Value) map[string]reflect.Value {
	keys := sr.MapKeys()
	m := make(map[string]reflect.Value, len(keys))
	for _, key := range keys {
		m[key.String()] = sr.MapIndex(key)
	}
	return m
}

func ConvertFromPbParams(c eval.Context, params []*fsmpb.Parameter) []eval.Parameter {
	ps := make([]eval.Parameter, len(params))
	for i, p := range params {
		ld := p.GetLookup()
		var lookup eval.Value
		if ld != nil {
			lookup = proto.FromPBData(ld)
		}
		ps[i] = impl.NewParameter(p.GetName(), c.ParseType2(p.GetType()), lookup, false)
	}
	return ps
}

func ConvertToPbParams(params []eval.Parameter) []*fsmpb.Parameter {
	ps := make([]*fsmpb.Parameter, len(params))
	for i, p := range params {
		ps[i] = &fsmpb.Parameter{Name: p.Name(), Type: p.Type().String()}
	}
	return ps
}

func ConvertIterate(def *datapb.Data) eval.Value {
	if def != nil {
		return proto.FromPBData(def)
	}
	return nil
}
