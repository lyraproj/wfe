package wfe

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-issues/issue"
	"net/url"
)

type resource struct {
	Activity
	typ eval.ObjectType
	extId string
	state eval.KeyedValue
}

func Resource(c eval.Context, n string, resourceType eval.ObjectType, state eval.KeyedValue, externalId string, input, output []eval.Parameter, w api.Condition) api.Activity {
	r := &resource{typ: resourceType, state: state, extId: externalId}
	r.Init(n, input, output, w)
	return r
}

func (r *resource) Identifier() string {
	vs := make(url.Values, 3)
	vs.Add(`resource_type`, r.typ.String())
	if r.extId != `` {
		vs.Add(`external_id`, r.extId)
	}
	return r.Activity.Identifier() + `?` + vs.Encode()
}

func (r *resource) Run(ctx eval.Context, input eval.KeyedValue) eval.KeyedValue {
	return ctx.Scope().WithLocalScope(func() (v eval.PValue) {
		scope := ctx.Scope()
		input.EachPair(func(k, v eval.PValue) {
			scope.Set(k.String(), v)
		})
		resolvedState := types.ResolveDeferred(ctx, r.state)
		obj := eval.New(ctx, r.typ, resolvedState).(eval.PuppetObject)
		newState := Apply(ctx, r.Identifier(), obj)
		output := r.Output()
		entries := make([]*types.HashEntry, len(output))
		for i, o := range output {
			entries[i] = r.getValue(o, newState)
		}
		return types.WrapHash(entries)
	}).(eval.KeyedValue)
}

func (r *resource) Label() string {
	return ActivityLabel(r)
}

func (r *resource) Style() string {
	return `resource`
}

func (r *resource) getValue(p eval.Parameter, o eval.PuppetObject) *types.HashEntry {
	n := p.Name()
	a := n
	v := p.Value()
	if a, ok := v.(*types.ArrayValue); ok {
		// Build hash from multiple attributes
		entries := make([]*types.HashEntry, a.Len())
		a.EachWithIndex(func(e eval.PValue, i int) {
			a := e.String()
			if v, ok := o.Get(a); ok {
				entries[i] = types.WrapHashEntry(e, v)
			} else {
				panic(eval.Error(WF_NO_SUCH_ATTRIBUTE, issue.H{`activity`: r, `name`: a}))
			}
		})
		return types.WrapHashEntry2(n, types.WrapHash(entries))
	}

	if s, ok := v.(*types.StringValue); ok {
		a = s.String()
	}
	if v, ok := o.Get(a); ok {
		return types.WrapHashEntry2(n, v)
	}
	panic(eval.Error(WF_NO_SUCH_ATTRIBUTE, issue.H{`activity`: r, `name`: a}))
}