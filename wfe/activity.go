package wfe

import (
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/wfe/condition"
	"github.com/puppetlabs/go-issues/issue"
	"net/url"
	"strings"
)

type Activity struct {
	name   string
	when   api.Condition
	input  []eval.Parameter
	output []eval.Parameter
}

func ActivityContext(c eval.Context) eval.OrderedMap {
	if ac, ok := c.Scope().Get(`genesis::context`); ok {
		return eval.AssertInstance(`invalid activity context`, types.DefaultHashType(), ac).(eval.OrderedMap)
	}
	panic(eval.Error(api.WF_NO_ACTIVITY_CONTEXT, issue.NO_ARGS))
}

func LeafName(activity api.Activity) string {
	names := strings.Split(activity.Name(), `::`)
	return names[len(names)-1]
}

func (a *Activity) Style() string {
	return `activity`
}

func ActivityLabel(a api.Activity) string {
	return fmt.Sprintf(`%s '%s'`, a.Style(), a.Name())
}

func (a *Activity) When() api.Condition {
	return a.when
}

func (a *Activity) Name() string {
	return a.name
}

func (a *Activity) Input() []eval.Parameter {
	return a.input
}

func (a *Activity) Output() []eval.Parameter {
	return a.output
}

func (a *Activity) Init(name string, input, output []eval.Parameter, when api.Condition) {
	if input == nil {
		input = eval.NoParameters
	}
	if output == nil {
		output = eval.NoParameters
	}
	if when == nil {
		when = condition.Always
	}
	a.name = name
	a.input = input
	a.output = output
	a.when = when
}

func (a *Activity) Identifier() string {
	return `genesis://puppet.com/` + a.Style() + `/` + url.PathEscape(a.name)
}

func ResolveInput(ctx eval.Context, a api.Activity, input eval.OrderedMap, p eval.Parameter) eval.Value {
	v := p.Value()
	if v == nil {
		var ok bool
		if v, ok = input.Get4(p.Name()); !ok {
			panic(eval.Error(WF_PARAMETER_UNRESOLVED, issue.H{`activity`: a, `parameter`: p.Name()}))
		}
	}
	return types.ResolveDeferred(ctx, v)
}
