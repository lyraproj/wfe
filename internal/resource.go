package internal

import (
	"net/url"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/wfe"
)

type resource struct {
	step
	typ     px.ObjectType
	handler px.TypedName
	extId   px.Value
}

func (r *resource) Type() px.ObjectType {
	return r.typ
}

func (r *resource) HandlerId() px.TypedName {
	return r.handler
}

func (r *resource) ExtId() px.Value {
	return r.extId
}

func newResource(c px.Context, def serviceapi.Definition) wfe.Step {
	r := &resource{}
	r.initStep(def)
	if eid, ok := wfe.GetOptionalProperty(def, `externalId`, types.DefaultStringType()); ok {
		r.extId = eid
	}

	rt := wfe.GetProperty(def, `resourceType`, types.DefaultTypeType()).(px.Type)
	if rs, ok := rt.(px.ResolvableType); ok {
		// Ensure that the handler for the resource type is loaded prior to attempting
		// the resolve.
		if tr, ok := rs.(*types.TypeReferenceType); ok && types.TypeNamePattern.MatchString(tr.TypeString()) {
			if _, ok = px.Load(c, px.NewTypedName(px.NsHandler, tr.TypeString())); ok {
				rt = rs.Resolve(c)
			}
		}
	}
	r.typ = px.AssertType(func() string { return "property resourceType of step " + def.Identifier().Name() },
		types.DefaultObjectType(), rt).(px.ObjectType)
	r.handler = px.NewTypedName(px.NsHandler, r.typ.Name())
	return r
}

func (r *resource) Identifier() string {
	return StepId(r)
}

func (r *resource) IdParams() url.Values {
	vs := r.step.IdParams()
	vs.Add(`rt`, r.typ.Name())
	vs.Add(`hid`, r.HandlerId().Name())
	return vs
}

func (r *resource) Run(c px.Context, parameters px.OrderedMap) px.OrderedMap {
	return applyState(c, r, parameters)
}

func (r *resource) Label() string {
	return StepLabel(r)
}

func (r *resource) Style() string {
	return `resource`
}

func (r *resource) WithIndex(index int) wfe.Step {
	rc := *r // Copy by value
	rc.setIndex(index)
	return &rc
}
