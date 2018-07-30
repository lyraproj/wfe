package server

import (
	"fmt"
	"golang.org/x/net/context"
)

func ExampleGenesis_Run() {
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

	ctx := NewContext(context.Background())
	ctx.Action("a", func(g Context) (*OutA, error) {
		return &OutA{`hello`, 4}, nil
	})

	ctx.Action("b1", func(g Context, in *InB) (*OutB1, error) {
		return &OutB1{in.A + ` world`, in.B + 4}, nil
	})

	ctx.Action("b2", func(g Context, in *InB) (*OutB2, error) {
		return &OutB2{in.A + ` earth`, in.B + 8}, nil
	})

	ctx.Action("c", func(g Context, in *InC) error {
		fmt.Printf("%s, %d, %s, %d\n", in.C, in.D, in.E, in.F)
		return nil
	})

	err := ctx.Validate()
	if err == nil {
		err = ctx.Run()
	}
	if err != nil {
		fmt.Println(err)
	}

	// Output: hello world, 8, hello earth, 12
}
