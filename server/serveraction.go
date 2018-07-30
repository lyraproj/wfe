package server

import (
	"github.com/puppetlabs/go-fsm/fsm"
	"gonum.org/v1/gonum/graph"
)

type ServerAction interface {
	fsm.Action
	graph.Node

	SetResolved()

	// Resolved channel will be closed when the action is resolved
	Resolved() <-chan bool
}

type serverAction struct {
	fsm.Action
	graph.Node
	resolved chan bool
}

func (a *serverAction) SetResolved() {
	close(a.resolved)
}

func (a *serverAction) Resolved() <-chan bool {
	return a.resolved
}
