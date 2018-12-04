package wfe

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/servicesdk/serviceapi"
)

type stateless struct {
	Activity
	api eval.ObjectType
}

var ioType = types.NewHashType(types.DefaultStringType(), types.DefaultRichDataType(), nil)
var expectedType = types.NewCallableType(
	types.NewTupleType([]eval.Type{ioType}, nil), ioType, nil)

func Stateless(def serviceapi.Definition) api.Activity {
	a := &stateless{}
	a.Init(def)
	return a
}

func (s *stateless) Init(d serviceapi.Definition) {
	s.Activity.Init(d)
	if api, ok := d.Properties().Get4(`interface`); ok {
		s.api = api.(eval.ObjectType)
	}
}

func (s *stateless) Run(ctx eval.Context, input eval.OrderedMap) eval.OrderedMap {
	service := s.GetService(ctx)
	result := service.Invoke(ctx, s.Name(), `do`, input)
	if m, ok := result.(eval.OrderedMap); ok {
		return m
	}
	panic(result.String())
}

func (s *stateless) Label() string {
	return ActivityLabel(s)
}

func (s *stateless) Style() string {
	return `stateless`
}
