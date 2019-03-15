package service

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/annotation"
	"github.com/lyraproj/servicesdk/wf"
	"github.com/lyraproj/wfe/api"
)

func StartEra(c px.Context) {
	getIdentity(c).bumpEra(c)
}

// SweepAndGC performs a sweep of the Identity store, retrieves all garbage, and
// then tells the handler for each garbage entry to delete the resource. The entry
// is then purged from the Identity store
func SweepAndGC(c px.Context, prefix string) {
	identity := getIdentity(c)
	identity.sweep(c, prefix)
	gl := identity.garbage(c)
	ng := gl.Len()
	rs := make([]px.List, ng)

	// Store in reverse order
	ng--
	gl.EachWithIndex(func(t px.Value, i int) {
		rs[ng-i] = t.(px.List)
	})

	for _, l := range rs {
		hid := types.ParseURI(l.At(0).String()).Query().Get(`hid`)
		handlerDef := GetHandler(c, px.NewTypedName(px.NsHandler, hid))
		handler := GetService(c, handlerDef.ServiceId())

		extId := l.At(1)
		handler.Invoke(c, handlerDef.Identifier().Name(), `delete`, extId)
		identity.purgeExternal(c, extId)
	}
}

func ApplyState(c px.Context, resource api.Resource, input px.OrderedMap) px.OrderedMap {
	ac := ActivityContext(c)
	op := GetOperation(ac)

	handlerDef := GetHandler(c, resource.HandlerId())
	crd := GetProperty(handlerDef, `interface`, types.NewTypeType(types.DefaultObjectType())).(px.ObjectType)
	identity := getIdentity(c)
	handler := GetService(c, handlerDef.ServiceId())

	intId := types.WrapString(resource.Identifier())
	extId := resource.ExtId()
	explicitExtId := extId != nil
	if !explicitExtId {
		// external id must exist in order to do a read or delete
		extId = identity.getExternal(c, intId, op == wf.Read || op == wf.Delete)
	}

	var result px.PuppetObject
	hn := handlerDef.Identifier().Name()
	switch op {
	case wf.Read:
		if extId == nil {
			return px.EmptyMap
		}
		result = px.AssertInstance(handlerDef.Label, resource.Type(), handler.Invoke(c, hn, `read`, extId)).(px.PuppetObject)

	case wf.Upsert:
		if explicitExtId {
			// An explicit externalId is for resources not managed by us. Only possible action
			// here is a read
			result = handler.Invoke(c, hn, `read`, extId).(px.PuppetObject)
			break
		}

		desiredState := GetService(c, resource.ServiceId()).State(c, resource.Name(), input)
		if extId == nil {
			// Nothing exists yet. Create a new instance
			rt := handler.Invoke(c, hn, `create`, desiredState).(px.List)
			result = px.AssertInstance(handlerDef.Label, resource.Type(), rt.At(0)).(px.PuppetObject)
			extId = rt.At(1)
			identity.associate(c, intId, extId)
			break
		}

		// Read current state and check if an update is needed
		var updateNeeded, recreateNeeded bool
		currentState := px.AssertInstance(handlerDef.Label, resource.Type(), handler.Invoke(c, hn, `read`, extId)).(px.PuppetObject)

		if a, ok := resource.Type().Annotations(c).Get(annotation.ResourceType); ok {
			ra := a.(annotation.Resource)
			updateNeeded, recreateNeeded = ra.Changed(desiredState, currentState)
		} else {
			updateNeeded = !desiredState.Equals(currentState, nil)
			recreateNeeded = false
		}

		if updateNeeded {
			if !recreateNeeded {
				// Update existing content. If an update method exists, call it. If not, then fall back
				// to delete + create
				if _, ok := crd.Member(`update`); ok {
					result = px.AssertInstance(handlerDef.Label, resource.Type(), handler.Invoke(c, hn, `update`, extId, desiredState)).(px.PuppetObject)
					break
				}
			}

			// Rely on that deletion happens by means of GC at end of run
			identity.removeExternal(c, extId)

			rt := handler.Invoke(c, hn, `create`, desiredState)
			rl := rt.(px.List)
			result = px.AssertInstance(handlerDef.Label, resource.Type(), rl.At(0)).(px.PuppetObject)
			extId = rl.At(1)
			identity.associate(c, intId, extId)
		} else {
			result = currentState
		}
	default:
		panic(px.Error(wf.IllegalOperation, issue.H{`operation`: op}))
	}

	switch op {
	case wf.Read, wf.Upsert:
		output := resource.Output()
		entries := make([]*types.HashEntry, len(output))
		for i, o := range output {
			entries[i] = getValue(o, resource, result)
		}
		return types.WrapHash(entries)
	}
	return px.EmptyMap
}

func getValue(p px.Parameter, r api.Resource, o px.PuppetObject) *types.HashEntry {
	n := p.Name()
	a := n
	if p.HasValue() {
		v := p.Value()
		if a, ok := v.(*types.Array); ok {
			// Build hash from multiple attributes
			entries := make([]*types.HashEntry, a.Len())
			a.EachWithIndex(func(e px.Value, i int) {
				a := e.String()
				if v, ok := o.Get(a); ok {
					entries[i] = types.WrapHashEntry(e, v)
				} else {
					panic(px.Error(api.NoSuchAttribute, issue.H{`activity`: r, `name`: a}))
				}
			})
			return types.WrapHashEntry2(n, types.WrapHash(entries))
		}

		if s, ok := v.(px.StringValue); ok {
			a = s.String()
		}
	}
	if v, ok := o.Get(a); ok {
		return types.WrapHashEntry2(n, v)
	}
	panic(px.Error(api.NoSuchAttribute, issue.H{`activity`: r, `name`: a}))
}
