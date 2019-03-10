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

func Resource(def serviceapi.Definition) api.Activity {
	r := &resource{}
	r.Init(def)
	return r
}

func (r *resource) Init(d serviceapi.Definition) {
	r.Activity.Init(d)
	if eid, ok := service.GetOptionalProperty(d, `externalId`, types.DefaultStringType()); ok {
		r.extId = eid
	}
	r.typ = service.GetProperty(d, `resourceType`, types.NewTypeType(types.DefaultObjectType())).(px.ObjectType)
	r.handler = px.NewTypedName(px.NsHandler, r.typ.Name())
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
