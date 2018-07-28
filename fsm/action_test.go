package fsm

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"fmt"

	// Initialize puppet core
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

func ExampleGenesis_Action() {
	type StartActionResult struct {
		A string
		B int64
	}

	type ActionAResult struct {
		C string
		D int64
	}

	type ActionBResult struct {
		E string
		F int64
	}

	err := eval.Puppet.Do(func(ctx eval.Context) error {
		genesis := GetGenesisService(ctx)
		genesis.Action("a", func(g Genesis) (*StartActionResult, error) {
			return &StartActionResult{`hello`, 4}, nil
		})

		genesis.Action("b1", func(g Genesis, a string, b int64) (*ActionAResult, error) {
			return &ActionAResult{a + ` world`, b + 4}, nil
		}, `a.a`, `a.b`)

		genesis.Action("b2", func(g Genesis, a string, b int64) (*ActionBResult, error) {
			return &ActionBResult{a + ` earth`, b + 8}, nil
		}, `a.a`, `a.b`)

		genesis.Action("c", func(g Genesis, c string, d int64, e string, f int64) error {
			fmt.Printf("%s, %d, %s, %d\n", c, d, e, f)
			return nil
		}, `b1.c`, `b1.d`, `b2.e`, `b2.f`)

		err := genesis.(GenesisService).Validate()
		if err == nil {
			err = genesis.(GenesisService).Run()
		}
		return err
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output: hello world, 8, hello earth, 12
}

