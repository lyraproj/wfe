package wfe

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wfapi"
	"net/url"
)

type resource struct {
	Activity
	typ   eval.ObjectType
	handler eval.TypedName
	extId eval.Value
}

func Resource(def serviceapi.Definition) api.Activity {
	r := &resource{}
	r.Init(def)
	return r
}

func (r *resource) Init(d serviceapi.Definition) {
	r.Activity.Init(d)
	if eid, ok := GetOptionalProperty(d, `external_id`, types.DefaultStringType()); ok {
		r.extId = eid
	}
	r.typ = GetProperty(d, `resource_type`, types.NewTypeType(types.DefaultObjectType())).(eval.ObjectType)
	r.handler = eval.NewTypedName(eval.NsHandler, r.typ.Name())
}

func (r *resource) Identifier() string {
	vs := make(url.Values, 3)
	vs.Add(`resource_type`, r.typ.Name())
	if r.extId != nil {
		vs.Add(`external_id`, r.extId.String())
	}
	return r.Activity.Identifier() + `?` + vs.Encode()
}

func (r *resource) Run(c eval.Context, input eval.OrderedMap) eval.OrderedMap {
	ac := ActivityContext(c)
	op := GetOperation(ac)

	handlerDef := GetHandler(c, r.handler)
	crd := GetProperty(handlerDef, `interface`, types.NewTypeType(types.DefaultObjectType())).(eval.ObjectType)
	identity := getIdentity(c)
	handler := GetService(c, handlerDef.ServiceId())

	extId := r.extId
	explicitExtId := extId != nil
	if !explicitExtId {
		// external id must exist in order to do a read or delete
		if s, ok := identity.getExternal(c, r.Identifier(), op == wfapi.Read || op == wfapi.Delete); ok {
			extId = types.WrapString(s)
		}
	}

	var result eval.Value
	hn := handlerDef.Identifier().Name()
	switch op {
	case wfapi.Read:
		if extId == nil {
			return eval.EMPTY_MAP
		}
		result = handler.Invoke(c, hn, `read`, extId).(eval.OrderedMap)

	case wfapi.Upsert:
		if explicitExtId {
			// An explicit externalId is for resources not managed by us. Only possible action
			// here is a read
			result = handler.Invoke(c, hn, `read`, extId).(eval.OrderedMap)
			break
		}

		desiredState := r.GetService(c).State(c, r.name, input)
		if extId == nil {
			// Nothing exists yet. Create a new instance
			rt := handler.Invoke(c, hn, `create`, desiredState).(eval.List)
			result = rt.At(0)
			extId = rt.At(1)
			identity.associate(c, r.Identifier(), extId.String())
			break
		}

		// Update existing content. If an update method exists, call it. If not, then fall back
		// to delete + create
		if _, ok := crd.Member(`update`); ok {
			result = handler.Invoke(c, hn, `update`, extId, desiredState).(eval.OrderedMap)
			break
		}
		handler.Invoke(c, hn, `delete`, extId)
		identity.removeExternal(c, extId.String())

		rt := handler.Invoke(c, hn, `create`, desiredState)
		rl := rt.(eval.List)
		result = rl.At(0)
		extId = rl.At(1)
		identity.associate(c, r.Identifier(), extId.String())

	case wfapi.Delete:
		if !explicitExtId {
			handler.Invoke(c, hn, `delete`, extId)
			identity.removeExternal(c, extId.String())
		}
		return eval.EMPTY_MAP
	default:
		panic(eval.Error(wfapi.WF_ILLEGAL_OPERATION, issue.H{`operation`: op}))
	}

	switch op {
	case wfapi.Read, wfapi.Upsert:
		if newState, ok := result.(eval.PuppetObject); ok {
			eval.AssertInstance(handlerDef.Label, r.typ, newState)
			output := r.Output()
			entries := make([]*types.HashEntry, len(output))
			for i, o := range output {
				entries[i] = r.getValue(o, newState)
			}
			return types.WrapHash(entries)
		}
		c.Logger().Log(eval.INFO, result)
		panic(eval.Error(WF_OPERATION_DID_NOT_RETURN_STATE, issue.H{`op`: op.String(), `handler`: handlerDef}))
	}
	return eval.EMPTY_MAP
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
	if p.HasValue() {
		v := p.Value()
		if a, ok := v.(*types.ArrayValue); ok {
			// Build hash from multiple attributes
			entries := make([]*types.HashEntry, a.Len())
			a.EachWithIndex(func(e eval.Value, i int) {
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
	}
	if v, ok := o.Get(a); ok {
		return types.WrapHashEntry2(n, v)
	}
	panic(eval.Error(WF_NO_SUCH_ATTRIBUTE, issue.H{`activity`: r, `name`: a}))
}
