package wfe

import (
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-issues/issue"
	"sync/atomic"
)

const maxParallel = 100

type iterator struct {
	api.Activity
	over       []eval.Parameter
	variables  []eval.Parameter
	resultName string
}

func Iterator(style api.IterationStyle, activity api.Activity, name string, over, variables []eval.Parameter) api.Activity {
	switch style {
	case api.IterationStyleRange:
		return NewRange(activity, name, over, variables)
	case api.IterationStyleTimes:
		return NewTimes(activity, name, over, variables)
	default:
		panic(eval.Error(api.WF_ILLEGAL_ITERATION_STYLE, issue.H{`style`: style.String()}))
	}
}

func (it *iterator) IterationStyle() api.IterationStyle {
	panic("implement me")
}

func (it *iterator) Over() []eval.Parameter {
	return it.over
}

func (it *iterator) Producer() api.Activity {
	return it.Activity
}

// Input returns the Input declared for the action + Over() and - Variables
func (it *iterator) Input() []eval.Parameter {
	input := it.Producer().Input()
	all := make([]eval.Parameter, 0, len(it.over)+len(input)-len(it.variables))
	all = append(all, it.over...)
nextInput:
	for _, in := range input {
		for _, v := range it.variables {
			if in.Name() == v.Name() {
				continue nextInput
			}
		}
		all = append(all, in)
	}
	return all
}

// Output returns the on parameter, named after the activity which is a hash of
// key and value parameters of the activity.
func (it *iterator) Output() []eval.Parameter {
	output := it.Producer().Output()

	// Constructor validates that activity output consists of two values, a key
	// and a value.
	key := output[0]
	value := output[1]
	return []eval.Parameter{
		impl.NewParameter(it.resultName, types.NewHashType(key.ValueType(), value.ValueType(), nil), nil, false)}
}

func (it *iterator) Variables() []eval.Parameter {
	return it.variables
}

func (it *iterator) iterateRange(ctx eval.Context, vars eval.OrderedMap, start, end int64) eval.OrderedMap {

	done := make(chan bool)
	count := end - start
	numWorkers := int(count)
	if count > maxParallel {
		numWorkers = maxParallel
	}

	entries := make([]*types.HashEntry, count)
	p0 := types.WrapString(it.Variables()[0].Name())
	p := it.Producer()
	jobs := make(chan int64)
	for i := 0; i < numWorkers; i++ {
		eval.Fork(ctx, func(fc eval.Context) {
			for ix := range jobs {
				func() {
					defer func() {
						if atomic.AddInt64(&count, -1) <= 0 {
							close(jobs)
							done <- true
						}
					}()
					result := p.Run(fc, vars.Merge(types.SingletonHash(p0, types.WrapInteger(int64(ix)))))
					entries[ix-start] = types.WrapHashEntry(result.Get5(`key`, eval.UNDEF), result.Get5(`value`, eval.UNDEF))
				}()
			}
		})
	}

	for i := start; i < end; i++ {
		jobs <- i
	}
	<-done
	return types.SingletonHash2(it.resultName, types.WrapHash(entries))
}

func resolveInput(c eval.Context, it api.Iterator, input eval.OrderedMap) ([]eval.Value, eval.OrderedMap) {
	// Resolve the parameters that acts as input to the iteration.
	over := make([]eval.Value, len(it.Over()))
	vars := make([]*types.HashEntry, 0, len(it.Input())-len(it.Over()))

	for i, o := range it.Over() {
		arg := input.Get5(o.Name(), eval.UNDEF)
		if df, ok := arg.(types.Deferred); ok {
			arg = df.Resolve(c)
		}
		over[i] = arg
	}

	// Strip input intended for the iterator from the list intended for the activity that will be called by the iterator
nextInput:
	for _, ap := range it.Producer().Input() {
		for _, op := range it.Over() {
			if op.Name() == ap.Name() {
				continue nextInput
			}
		}
		if ev, ok := input.GetEntry(ap.Name()); ok {
			vars = append(vars, ev.(*types.HashEntry))
		}
	}
	return over, types.WrapHash(vars)
}

func assertOverCount(it api.Iterator, expectedCount int) {
	actualCount := len(it.Over())
	if actualCount != expectedCount {
		panic(eval.Error(WF_ITERATION_PARAMETER_INVALID_COUNT,
			issue.H{`iterator`: it, `expected`: expectedCount, `actual`: actualCount}))
	}
}

func assertVariableCount(it api.Iterator, expectedCount int) {
	actualCount := len(it.Variables())
	if actualCount != expectedCount {
		panic(eval.Error(WF_ITERATION_VARIABLE_INVALID_COUNT,
			issue.H{`iterator`: it, `expected`: expectedCount, `actual`: actualCount}))
	}
}

func Validate(it api.Iterator, expectedOver, expectedVars int) {
	assertOverCount(it, expectedOver)
	assertVariableCount(it, expectedVars)

	a := it.Producer()

	// Ensure that output consists of a key and a value parameter
	o := a.Output()
	if len(o) != 2 || !(o[0].Name() == `key` && o[1].Name() == `value` || o[1].Name() == `key` && o[0].Name() == `value`) {
		panic(eval.Error(WF_ITERATION_ACTIVITY_WRONG_OUTPUT, issue.H{`iterator`: it}))
	}

	// Ensure that input contains output produced by the iterator
	is := a.Input()
	vs := it.Variables()
nextVar:
	for _, v := range vs {
		for _, i := range is {
			if i.Name() == v.Name() {
				continue nextVar
			}
		}
		panic(eval.Error(WF_ITERATION_ACTIVITY_WRONG_INPUT, issue.H{`iterator`: it}))
	}
}

func assertInt(t api.Iterator, arg eval.Value, paramIdx int) int64 {
	iv, ok := arg.(*types.IntegerValue)
	if !ok {
		panic(eval.Error(WF_ITERATION_PARAMETER_WRONG_TYPE, issue.H{
			`iterator`: t, `parameter`: t.Over()[paramIdx].Name(), `expected`: `Integer`, `actual`: arg.Type()}))
	}
	return iv.Int()
}

func iterLabel(it api.Iterator) string {
	return fmt.Sprintf(`%s %s iteration`, it.Style(), ActivityLabel(it))
}

func NewTimes(activity api.Activity, name string, over []eval.Parameter, variables []eval.Parameter) api.Iterator {
	it := &times{iterator{activity, over, variables, name}}
	Validate(it, 1, 1)
	return it
}

type times struct {
	iterator
}

func (t *times) Label() string {
	return iterLabel(t)
}

func (t *times) IterationStyle() api.IterationStyle {
	return api.IterationStyleTimes
}

func (t *times) Run(ctx eval.Context, input eval.OrderedMap) eval.OrderedMap {
	over, vars := resolveInput(ctx, t, input)
	return t.iterateRange(ctx, vars, 0, assertInt(t, over[0], 0))
}

type itRange struct {
	iterator
}

func NewRange(activity api.Activity, name string, over []eval.Parameter, variables []eval.Parameter) api.Iterator {
	it := &itRange{iterator{activity, over, variables, name}}
	Validate(it, 2, 1)
	return it
}

func (t *itRange) Label() string {
	return iterLabel(t)
}

func (t *itRange) IterationStyle() api.IterationStyle {
	return api.IterationStyleRange
}

func (t *itRange) Run(ctx eval.Context, input eval.OrderedMap) eval.OrderedMap {
	over, vars := resolveInput(ctx, t, input)
	return t.iterateRange(ctx, vars, assertInt(t, over[0], 0), assertInt(t, over[1], 1))
}
