package shared

import (
	"reflect"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/data-protobuf/datapb"
)

func ExpandStringMap(sr reflect.Value) map[string]reflect.Value {
	keys := sr.MapKeys()
	m := make(map[string]reflect.Value, len(keys))
	for _, key := range keys {
		m[key.String()] = sr.MapIndex(key)
	}
	return m
}

func ConvertFromPbParams(params []*fsmpb.Parameter) []api.Parameter {
	ps := make([]api.Parameter, len(params))
	for i, p := range params {
		ld := p.GetLookup()
		var lookup *reflect.Value
		if ld != nil {
			lv, err := datapb.FromData(ld)
			if err != nil {
				panic(err)
			}
			lookup = &lv
		}
		ps[i] = api.NewParameter(p.GetName(), p.GetType(), lookup)
	}
	return ps
}

func ConvertToPbParams(params []api.Parameter) []*fsmpb.Parameter {
	ps := make([]*fsmpb.Parameter, len(params))
	for i, p := range params {
		ps[i] = &fsmpb.Parameter{Name: p.Name(), Type: p.Type()}
	}
	return ps
}
