package wfe

import (
	"fmt"
	"sync/atomic"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wfapi"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/wfe/service"
)

const maxParallel = 100

type iterator struct {
	api.Activity
	over       []px.Parameter
	variables  []px.Parameter
	resultName string
}

func Iterator(def serviceapi.Definition) api.Activity {
	over := getParameters(`over`, def.Properties())
	variables := getParameters(`variables`, def.Properties())
	style := wfapi.NewIterationStyle(service.GetStringProperty(def, `iterationStyle`))
	activity := CreateActivity(service.GetProperty(def, `producer`, serviceapi.DefinitionMetaType).(serviceapi.Definition))
	resultName := wfapi.LeafName(def.Identifier().Name())
	switch style {
	case wfapi.IterationStyleRange:
		return NewRange(activity, resultName, over, variables)
	case wfapi.IterationStyleTimes:
		return NewTimes(activity, resultName, over, variables)
	default:
		panic(px.Error(api.WF_ILLEGAL_ITERATION_STYLE, issue.H{`style`: style.String()}))
	}
}

func (it *iterator) IterationStyle() wfapi.IterationStyle {
	panic("implement me")
}

func (it *iterator) Over() []px.Parameter {
	return it.over
}

func (it *iterator) Producer() api.Activity {
	return it.Activity
}

// Input returns the Input declared for the stateHandler + Over() and - Variables
func (it *iterator) Input() []px.Parameter {
	input := it.Producer().Input()
	all := make([]px.Parameter, 0, len(it.over)+len(input)-len(it.variables))
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
func (it *iterator) Output() []px.Parameter {
	output := it.Producer().Output()

	// Constructor validates that activity output consists of two values, a key
	// and a value.
	key := output[0]
	value := output[1]
	return []px.Parameter{
		px.NewParameter(it.resultName, types.NewHashType(key.Type(), value.Type(), nil), nil, false)}
}

func (it *iterator) Variables() []px.Parameter {
	return it.variables
}

func (it *iterator) iterateRange(ctx px.Context, vars px.OrderedMap, start, end int64) px.OrderedMap {

	done := make(chan bool)
	count := end - start
	numWorkers := int(count)
	if count > maxParallel {
		numWorkers = maxParallel
	}

	entries := make([]*types.HashEntry, count)
	p0 := it.Variables()[0].Name()
	p := it.Producer()
	jobs := make(chan int64)
	for i := 0; i < numWorkers; i++ {
		px.Fork(ctx, func(fc px.Context) {
			for ix := range jobs {
				func() {
					defer func() {
						if atomic.AddInt64(&count, -1) <= 0 {
							close(jobs)
							done <- true
						}
					}()
					result := p.Run(fc, vars.Merge(px.SingletonMap(p0, types.WrapInteger(int64(ix)))))
					entries[ix-start] = types.WrapHashEntry(result.Get5(`key`, px.Undef), result.Get5(`value`, px.Undef))
				}()
			}
		})
	}

	for i := start; i < end; i++ {
		jobs <- i
	}
	<-done
	return px.SingletonMap(it.resultName, types.WrapHash(entries))
}

func resolveInput(c px.Context, it api.Iterator, input px.OrderedMap) ([]px.Value, px.OrderedMap) {
	// Resolve the parameters that acts as input to the iteration.
	over := make([]px.Value, len(it.Over()))
	vars := make([]*types.HashEntry, 0, len(it.Input())-len(it.Over()))

	for i, o := range it.Over() {
		arg := input.Get5(o.Name(), px.Undef)
		if df, ok := arg.(types.Deferred); ok {
			arg = df.Resolve(c, c.Scope())
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
		panic(px.Error(WF_ITERATION_PARAMETER_INVALID_COUNT,
			issue.H{`iterator`: it, `expected`: expectedCount, `actual`: actualCount}))
	}
}

func assertVariableCount(it api.Iterator, expectedCount int) {
	actualCount := len(it.Variables())
	if actualCount != expectedCount {
		panic(px.Error(WF_ITERATION_VARIABLE_INVALID_COUNT,
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
		panic(px.Error(WF_ITERATION_ACTIVITY_WRONG_OUTPUT, issue.H{`iterator`: it}))
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
		panic(px.Error(WF_ITERATION_ACTIVITY_WRONG_INPUT, issue.H{`iterator`: it}))
	}
}

func assertInt(t api.Iterator, arg px.Value, paramIdx int) int64 {
	iv, ok := arg.(px.Integer)
	if !ok {
		panic(px.Error(WF_ITERATION_PARAMETER_WRONG_TYPE, issue.H{
			`iterator`: t, `parameter`: t.Over()[paramIdx].Name(), `expected`: `Integer`, `actual`: arg.PType()}))
	}
	return iv.Int()
}

func iterLabel(it api.Iterator) string {
	return fmt.Sprintf(`%s %s iteration`, it.Style(), ActivityLabel(it))
}

func NewTimes(activity api.Activity, name string, over []px.Parameter, variables []px.Parameter) api.Iterator {
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

func (t *times) IterationStyle() wfapi.IterationStyle {
	return wfapi.IterationStyleTimes
}

func (t *times) Run(ctx px.Context, input px.OrderedMap) px.OrderedMap {
	over, vars := resolveInput(ctx, t, input)
	return t.iterateRange(ctx, vars, 0, assertInt(t, over[0], 0))
}

type itRange struct {
	iterator
}

func NewRange(activity api.Activity, name string, over []px.Parameter, variables []px.Parameter) api.Iterator {
	it := &itRange{iterator{activity, over, variables, name}}
	Validate(it, 2, 1)
	return it
}

func (t *itRange) Label() string {
	return iterLabel(t)
}

func (t *itRange) IterationStyle() wfapi.IterationStyle {
	return wfapi.IterationStyleRange
}

func (t *itRange) Run(ctx px.Context, input px.OrderedMap) px.OrderedMap {
	over, vars := resolveInput(ctx, t, input)
	return t.iterateRange(ctx, vars, assertInt(t, over[0], 0), assertInt(t, over[1], 1))
}
