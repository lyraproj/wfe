package api

type Parameter interface {
	Name() string

	Type() string
}

type parameter struct {
	name string
	typ  string
}

func NewParameter(name, typ string) Parameter {
	return &parameter{name, typ}
}

func (p *parameter) Name() string {
	return p.name
}

func (p *parameter) Type() string {
	return p.typ
}
