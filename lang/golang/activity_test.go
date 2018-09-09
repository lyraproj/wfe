package golang

import (
	"context"
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/wfe"
	"github.com/puppetlabs/go-hiera/lookup"
	"strings"

	// Ensure Pcore and lookup are initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
	_ "github.com/puppetlabs/go-hiera/functions"
)

func ExampleActivity() {
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

	type OutC struct {
		R string
	}

	err := eval.Puppet.Do(func(ctx eval.Context) error {

		// Run actions in process by adding actions directly to the actor server
		as := wfe.NewWorkflow(`wftest`, nil, MakeParams(ctx, "wftest", &OutC{}), nil,

			Activity(ctx, "a", func() (*OutA, error) {
				return &OutA{`hello`, 4}, nil
			}),

			Activity(ctx, "b1", func(in *InB) (*OutB1, error) {
				return &OutB1{in.A + ` world`, in.B + 4}, nil
			}),

			Activity(ctx, "b2", func(in *InB) (*OutB2, error) {
				return &OutB2{in.A + ` earth`, in.B + 8}, nil
			}),

			Activity(ctx, "c", func(in *InC) (*OutC, error) {
				return &OutC{fmt.Sprintf("%s, %d, %s, %d\n", in.C, in.D, in.E, in.F)}, nil
			}))

		result := as.Run(ctx, eval.EMPTY_MAP)
		fmt.Println(result.Get5(`r`, eval.EMPTY_STRING))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output: hello world, 8, hello earth, 12
}

func ExampleActivity_failValiation() {
	type AB struct {
		A string
		B int64
	}

	type CD struct {
		C string
		D int64
	}

	type ABCDE struct {
		A string
		B int64
		C string
		D int64
		E string
	}

	type F struct {
		F string
	}

	err := eval.Puppet.Do(func(ctx eval.Context) error {

		// Run actions in process by adding actions directly to the actor server
		as := wfe.NewWorkflow(`wftest`, []eval.Parameter{}, []eval.Parameter{}, nil,

			Activity(ctx, "a", func() (*AB, error) {
				return &AB{`hello`, 4}, nil
			}),

			Activity(ctx, "b", func(in *AB) (*CD, error) {
				return &CD{in.A + ` world`, in.B + 4}, nil
			}),

			Activity(ctx, "c", func(in *ABCDE) (*F, error) {
				return &F{in.A + ` earth`}, nil
			}))

		result := as.Run(ctx, eval.EMPTY_MAP)
		fmt.Println(result.Get5(`r`, eval.EMPTY_STRING))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output: activity 'c' value 'e' is never produced
}

var sampleData = map[string]eval.PValue{
	`L`: types.WrapString(`value of L`),
}

func provider(c lookup.Context, key string, _ eval.KeyedValue) eval.PValue {
	if v, ok := sampleData[key]; ok {
		return v
	}
	c.NotFound()
	return nil
}

func ExampleActivity_lookup() {
	type L struct {
		L string `puppet:"value=>Deferred(lookup,['L'])"`
	}

	type AB struct {
		A string
		B int64
	}

	type CD struct {
		C string
		D int64
	}

	type ABCD struct {
		A string
		B int64
		C string
		D int64
	}

	type F struct {
		F string
	}

	err := lookup.DoWithParent(context.Background(), provider, func(ctx lookup.Context) error {
		// Run actions in process by adding actions directly to the actor server
		wf := wfe.NewWorkflow(`wftest`, []eval.Parameter{}, MakeParams(ctx, `wftest`, &F{}), nil,

			Activity(ctx, "a", func(in *L) (*AB, error) {
				return &AB{in.L, 4}, nil
			}),

			Activity(ctx, "b", func(in *AB) (*CD, error) {
				return &CD{strings.ToLower(in.A), in.B + 4}, nil
			}),

			Activity(ctx, "c", func(in *ABCD) (*F, error) {
				return &F{fmt.Sprintf(`%s %d, %s %d`, in.A, in.B, in.C, in.D)}, nil
			}))

		result := wf.Run(ctx, eval.EMPTY_MAP)
		fmt.Println(result.Get5(`f`, eval.EMPTY_STRING))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output: value of L 4, value of l 8
}
