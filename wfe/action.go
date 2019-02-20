package wfe

import (
	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/api"
)

type action struct {
	Activity
	api eval.ObjectType
}

var ioType = types.NewHashType(types.DefaultStringType(), types.DefaultRichDataType(), nil)
var expectedType = types.NewCallableType(
	types.NewTupleType([]eval.Type{ioType}, nil), ioType, nil)

func Action(def serviceapi.Definition) api.Activity {
	a := &action{}
	a.Init(def)
	return a
}

func (s *action) Init(d serviceapi.Definition) {
	s.Activity.Init(d)
	if api, ok := d.Properties().Get4(`interface`); ok {
		s.api = api.(eval.ObjectType)
	}
}

func (s *action) Run(ctx eval.Context, input eval.OrderedMap) eval.OrderedMap {
	service := s.GetService(ctx)
	hclog.Default().Debug(`executing action`, `name`, s.name)
	result := service.Invoke(ctx, s.Name(), `do`, input)
	if m, ok := result.(eval.OrderedMap); ok {
		return m
	}
	panic(result.String())
}

func (s *action) Label() string {
	return ActivityLabel(s)
}

func (s *action) Style() string {
	return `action`
}
