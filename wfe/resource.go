package wfe

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/wfe/service"
	"net/url"
)

type resource struct {
	Activity
	typ     eval.ObjectType
	handler eval.TypedName
	extId   eval.Value
}

func (r *resource) Type() eval.ObjectType {
	return r.typ
}

func (r *resource) HandlerId() eval.TypedName {
	return r.handler
}

func (r *resource) ExtId() eval.Value {
	return r.extId
}

func Resource(def serviceapi.Definition) api.Activity {
	r := &resource{}
	r.Init(def)
	return r
}

func (r *resource) Init(d serviceapi.Definition) {
	r.Activity.Init(d)
	if eid, ok := service.GetOptionalProperty(d, `external_id`, types.DefaultStringType()); ok {
		r.extId = eid
	}
	r.typ = service.GetProperty(d, `resource_type`, types.NewTypeType(types.DefaultObjectType())).(eval.ObjectType)
	r.handler = eval.NewTypedName(eval.NsHandler, r.typ.Name())
}

func (r *resource) Identifier() string {
	vs := make(url.Values, 3)
	vs.Add(`rt`, r.typ.Name())
	vs.Add(`hid`, r.HandlerId().Name())
	return r.Activity.Identifier() + `?` + vs.Encode()
}

func (r *resource) Run(c eval.Context, input eval.OrderedMap) eval.OrderedMap {
	return service.ApplyState(c, r, input)
}

func (r *resource) Label() string {
	return ActivityLabel(r)
}

func (r *resource) Style() string {
	return `resource`
}
