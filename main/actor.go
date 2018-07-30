package main

import (
	"fmt"
	"github.com/puppetlabs/go-fsm/fsm"
	"log"
	"github.com/puppetlabs/go-fsm/actor"
)

type OutA struct {
	A string
	B int64
}

type InB struct {
	A string
	B int64
}

type OutB1 struct {
	C string
	D int64
}

type OutB2 struct {
	E string
	F int64
}

type InC struct {
	C string
	D int64
	E string
	F int64
}

func main() {
	ctx := actor.NewContext()

	ctx.Action("a", func(g fsm.Context) (*OutA, error) {
		return &OutA{`hello`, 4}, nil
	})

	ctx.Action("b1", func(g fsm.Context, in *InB) (*OutB1, error) {
		return &OutB1{in.A + ` world`, in.B + 4}, nil
	})

	ctx.Action("b2", func(g fsm.Context, in *InB) (*OutB2, error) {
		return &OutB2{in.A + ` earth`, in.B + 8}, nil
	})

	ctx.Action("c", func(g fsm.Context, in *InC) error {
		fmt.Printf("%s, %d, %s, %d\n", in.C, in.D, in.E, in.F)
		return nil
	})

	err := ctx.RegisterServer("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to create fsmService: %v", err)
	}
}
