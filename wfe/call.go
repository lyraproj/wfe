package wfe

import (
	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wf"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/wfe/service"
)

type call struct {
	Step
	ra api.Step
}

func Call(c px.Context, def serviceapi.Definition) api.Step {
	r := &call{}
	r.Init(def)
	call := service.GetStringProperty(def, `call`)
	hclog.Default().Debug(`resolving step call`, `name`, r.name, `call`, call)
	r.ra = CreateStep(c, service.GetDefinition(c, px.NewTypedName(px.NsDefinition, call)))
	return r
}

func (r *call) CalledStep() api.Step {
	return r.ra
}

func (r *call) Identifier() string {
	return StepId(r)
}

func (r *call) Parameters() []serviceapi.Parameter {
	var input []serviceapi.Parameter
	if len(r.parameters) == 0 {
		input = r.ra.Parameters()
	} else {
		input = r.parameters
	}
	return input
}

func (r *call) Returns() []serviceapi.Parameter {
	output := r.returns
	if len(output) == 0 {
		output = r.ra.Returns()
	}
	return output
}

func (r *call) When() wf.Condition {
	when := r.when
	if when == nil {
		when = r.ra.When()
	} else {
		if r.ra.When() != nil {
			when = wf.And([]wf.Condition{when, r.ra.When()})
		}
	}
	return when
}

func (r *call) Run(ctx px.Context, input px.OrderedMap) px.OrderedMap {
	return r.mapOutput(r.ra.Run(ctx, r.mapInput(ResolveParameters(ctx, r, input))))
}

func (r *call) Label() string {
	return StepLabel(r)
}

func (r *call) Style() string {
	return `call`
}

func (r *call) mapInput(input px.OrderedMap) px.OrderedMap {
	ips := r.parameters
	if len(ips) == 0 {
		ips = r.ra.Parameters()
		if len(ips) == 0 {
			return input
		}
	}
	return input.MapEntries(func(entry px.MapEntry) px.MapEntry {
		key := entry.Key()
		kn := key.String()
		for _, p := range ips {
			if p.Name() == kn {
				if p.Alias() != `` {
					entry = types.WrapHashEntry2(p.Alias(), entry.Value())
				}
				break
			}
		}
		return entry
	})
}

func (r *call) mapOutput(output px.OrderedMap) px.OrderedMap {
	ops := r.Returns()
	if len(ops) == 0 {
		return output
	}
	return output.MapEntries(func(entry px.MapEntry) px.MapEntry {
		key := entry.Key().String()
		for _, p := range ops {
			if p.Alias() != `` {
				if p.Alias() == key {
					entry = types.WrapHashEntry2(p.Name(), entry.Value())
					break
				}
			}
		}
		return entry
	})
}

func (r *call) WithIndex(index int) api.Step {
	rc := *r // Copy by value
	rc.setIndex(index)
	return &rc
}
