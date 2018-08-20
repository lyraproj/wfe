package fsm

import (
	"context"
	"fmt"
	"github.com/puppetlabs/go-fsm/api"
	"reflect"
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
	as := NewActorServer(context.Background(), `test`, []api.Parameter{}, []api.Parameter{})
	as.Action("a", func(s api.Genesis) (*OutA, error) {
		return &OutA{`hello`, 4}, nil
	})

	as.Action("b1", func(g api.Genesis, in *InB) (*OutB1, error) {
		return &OutB1{in.A + ` world`, in.B + 4}, nil
	})

	as.Action("b2", func(g api.Genesis, in *InB) (*OutB2, error) {
		return &OutB2{in.A + ` earth`, in.B + 8}, nil
	})

	as.Action("c", func(g api.Genesis, in *InC) error {
		fmt.Printf("%s, %d, %s, %d\n", in.C, in.D, in.E, in.F)
		return nil
	})

	err := as.Validate()
	if err == nil {
		fmt.Println(as.Call(nil, map[string]reflect.Value{}))
	}
	if err != nil {
		fmt.Println(err)
	}

	// Output: hello world, 8, hello earth, 12
}
