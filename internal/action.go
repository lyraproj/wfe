package internal

import (
	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/wfe"
)

type action struct {
	step
	api px.ObjectType
}

/* TODO: Add type check using expectedType
var ioType = types.NewHashType(types.DefaultStringType(), types.DefaultRichDataType(), nil)
var expectedType = types.NewCallableType(
	types.NewTupleType([]px.Type{ioType}, nil), ioType, nil)
*/

func newAction(def serviceapi.Definition) wfe.Step {
	a := &action{}
	a.Init(def)
	return a
}

func (a *action) Init(d serviceapi.Definition) {
	a.step.initStep(d)
	if i, ok := d.Properties().Get4(`interface`); ok {
		a.api = i.(px.ObjectType)
	}
}

func (a *action) Run(ctx px.Context, parameters px.OrderedMap) px.OrderedMap {
	service := a.GetService(ctx)
	hclog.Default().Debug(`executing action`, `name`, a.name)
	result := service.Invoke(ctx, a.Name(), `do`, parameters)
	if m, ok := result.(px.OrderedMap); ok {
		return m
	}
	if _, ok := result.(*types.UndefValue); ok {
		return px.EmptyMap
	}
	panic(result.String())
}

func (a *action) Label() string {
	return StepLabel(a)
}

func (a *action) Identifier() string {
	return StepId(a)
}

func (a *action) Style() string {
	return `action`
}

func (a *action) WithIndex(index int) wfe.Step {
	ac := *a // Copy by value
	ac.setIndex(index)
	return &ac
}
