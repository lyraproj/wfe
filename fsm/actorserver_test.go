package fsm

import (
	"context"
	"fmt"
)

func ExampleActorServer_Action() {
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

	// Run actions in process by adding actions directly to the actor server
	as := NewActorServer(context.Background())
	as.Action("a", func(g ActorServer) (*OutA, error) {
		return &OutA{`hello`, 4}, nil
	})

	as.Action("b1", func(g ActorServer, in *InB) (*OutB1, error) {
		return &OutB1{in.A + ` world`, in.B + 4}, nil
	})

	as.Action("b2", func(g ActorServer, in *InB) (*OutB2, error) {
		return &OutB2{in.A + ` earth`, in.B + 8}, nil
	})

	as.Action("c", func(g ActorServer, in *InC) error {
		fmt.Printf("%s, %d, %s, %d\n", in.C, in.D, in.E, in.F)
		return nil
	})

	err := as.Validate()
	if err == nil {
		err = as.Run()
	}
	if err != nil {
		fmt.Println(err)
	}

	// Output: hello world, 8, hello earth, 12
}
