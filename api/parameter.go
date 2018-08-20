package api

import "reflect"

type Parameter interface {
	Name() string

	Type() string

	Lookup() *reflect.Value
}

type parameter struct {
	name string
	typ  string
	lookup *reflect.Value
}

func NewParameter(name, typ string, lookup *reflect.Value) Parameter {
	return &parameter{name, typ, lookup}
}

func (p *parameter) Name() string {
	return p.name
}

func (p *parameter) Lookup() *reflect.Value {
	return p.lookup
}

func (p *parameter) Type() string {
	return p.typ
}
