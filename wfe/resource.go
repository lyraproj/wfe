package wfe

import (
	"net/url"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/wfe/service"
)

type resource struct {
	Activity
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

func Resource(c px.Context, def serviceapi.Definition) api.Activity {
	r := &resource{}
	r.Init(def)
	if eid, ok := service.GetOptionalProperty(def, `externalId`, types.DefaultStringType()); ok {
		r.extId = eid
	}

	rt := service.GetProperty(def, `resourceType`, types.DefaultTypeType()).(px.Type)
	if rs, ok := rt.(px.ResolvableType); ok {
		// Ensure that the handler for the resource type is loaded prior to attempting
		// the resolve.
		if tr, ok := rs.(*types.TypeReferenceType); ok && types.TypeNamePattern.MatchString(tr.TypeString()) {
			_, ok = px.Load(c, px.NewTypedName(px.NsHandler, tr.TypeString()))
		}
		if ok {
			rt = rs.Resolve(c)
		}
	}
	r.typ = px.AssertType(func() string { return "property resourceType of activity " + def.Identifier().Name() },
		types.DefaultObjectType(), rt).(px.ObjectType)
	r.handler = px.NewTypedName(px.NsHandler, r.typ.Name())
	return r
}

func (r *resource) Identifier() string {
	vs := make(url.Values, 3)
	vs.Add(`rt`, r.typ.Name())
	vs.Add(`hid`, r.HandlerId().Name())
	return r.Activity.Identifier() + `?` + vs.Encode()
}

func (r *resource) Run(c px.Context, input px.OrderedMap) px.OrderedMap {
	return service.ApplyState(c, r, input)
}

func (r *resource) Label() string {
	return ActivityLabel(r)
}

func (r *resource) Style() string {
	return `resource`
}
