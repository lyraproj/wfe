package golang

import (
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-fsm/wfe"
	"strconv"
	// Ensure Pcore and lookup are initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
	_ "github.com/puppetlabs/go-hiera/functions"
)

func ExampleRange() {
	type R struct {
		From int64
		To   int64
	}

	type Entry struct {
		Key   string
		Value int64
	}

	err := eval.Puppet.Do(func(ctx eval.Context) error {
		// Run actions in process by adding actions directly to the actor server
		wf := wfe.NewWorkflow(`wftest`, MakeParams(ctx, `wftest`, &R{}), MakeParams(ctx, `wftest`, &struct {
			A map[string]int64
		}{}), nil,

			Range(ctx, "a", func(in *struct{ Idx int64 }) (*Entry, error) {
				return &Entry{strconv.Itoa(int(in.Idx)), in.Idx}, nil
			}, &R{}, &struct{ Idx int64 }{}))

		args := eval.Wrap(ctx, map[string]int64{`from`: 10, `to`: 15}).(eval.KeyedValue)
		result := wf.Run(ctx, args)
		fmt.Println(result.Get5(`a`, eval.EMPTY_STRING))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output: {'10' => 10, '11' => 11, '12' => 12, '13' => 13, '14' => 14}
}

func ExampleTimes() {
	type Count struct {
		Count int64
	}

	type Entry struct {
		Key   string
		Value int64
	}

	err := eval.Puppet.Do(func(ctx eval.Context) error {
		// Run actions in process by adding actions directly to the actor server
		as := wfe.NewWorkflow(`wftest`, MakeParams(ctx, `wftest`, &Count{}), MakeParams(ctx, `wftest`, &struct {
			A map[string]int64
		}{}), nil,

			Times(ctx, "a", func(in *struct{ Idx int64 }) (*Entry, error) {
				return &Entry{strconv.Itoa(int(in.Idx)), in.Idx}, nil
			}, &Count{}, &struct{ Idx int64 }{}))

		args := eval.Wrap(ctx, map[string]int64{`count`: 3}).(eval.KeyedValue)
		result := as.Run(ctx, args)
		fmt.Println(result.Get5(`a`, eval.EMPTY_STRING))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output: {'0' => 0, '1' => 1, '2' => 2}
}
