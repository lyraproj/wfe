package internal

import (
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/annotation"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wf"
	"github.com/lyraproj/wfe/wfe"
)

func readOrNotFound(c px.Context, identity serviceapi.Identity, handler serviceapi.Service, hn string, extId string) px.Value {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(issue.Reported); ok {
				if e, ok = e.Cause().(issue.Reported); ok && e.Code() == service.NotFound {
					// Not found by remote. Purge the extId and return nil.
					hclog.Default().Debug("Removing obsolete extId from Identity service", "extId", extId)
					identity.PurgeExternal(c, extId)
					return
				}
			}
			panic(r)
		}
	}()

	hclog.Default().Debug("Read state", "extId", extId)
	return handler.Invoke(c, hn, `read`, types.WrapString(extId))
}

func applyState(c px.Context, resource wfe.Resource, parameters px.OrderedMap) px.OrderedMap {
	ac := wfe.StepContext(c)
	op := wfe.GetOperation(ac)

	log := hclog.Default()
	handlerDef := wfe.GetHandler(c, resource.HandlerId())
	crd := wfe.GetProperty(handlerDef, `interface`, types.NewTypeType(types.DefaultObjectType())).(px.ObjectType)
	identity := GetLazyIdentity(c)
	handler := wfe.GetService(c, handlerDef.ServiceId())

	intId := resource.Identifier()
	extId := resource.ExtId()
	explicitExtId := extId != nil
	if !explicitExtId {
		// external id must exist in order to do a read or delete
		if eid, ok := identity.GetExternal(c, intId); ok {
			extId = types.WrapString(eid)
			log.Debug("GetExternal", "intId", intId, "extId", extId)
		} else if op == wf.Read || op == wf.Delete {
			panic(px.Error(wfe.UnableToDetermineExternalId, issue.H{`id`: intId}))
		}
	}

	var result px.PuppetObject
	hn := handlerDef.Identifier().Name()
	switch op {
	case wf.Read:
		if extId == nil {
			return px.EmptyMap
		}
		rt := readOrNotFound(c, identity, handler, hn, extId.String())
		if rt == nil {
			return px.EmptyMap
		}
		result = px.AssertInstance(handlerDef.Label, resource.Type(), rt).(px.PuppetObject)

	case wf.Upsert:
		if explicitExtId {
			// An explicit externalId is for resources not managed by us. Only possible action
			// here is a read
			rt := readOrNotFound(c, identity, handler, hn, extId.String())
			if rt == nil {
				// False positive from the Identity service
				return px.EmptyMap
			}
			result = px.AssertInstance(handlerDef.Label, resource.Type(), rt).(px.PuppetObject)
			break
		}

		desiredState := wfe.GetService(c, resource.ServiceId()).State(c, resource.Name(), parameters)
		if extId != nil {
			// Read current state and check if an update is needed
			rt := readOrNotFound(c, identity, handler, hn, extId.String())
			if rt == nil {
				// False positive from the Identity service
				extId = nil
			} else {
				result = px.AssertInstance(handlerDef.Label, resource.Type(), rt).(px.PuppetObject)
			}
		}

		if extId == nil {
			// Nothing exists yet. Create a new instance
			log.Debug("Create state", "intId", intId)
			rt := handler.Invoke(c, hn, `create`, desiredState).(px.List)
			result = px.AssertInstance(handlerDef.Label, resource.Type(), rt.At(0)).(px.PuppetObject)
			extId = rt.At(1)
			log.Debug("Associate state", "intId", intId, "extId", extId)
			identity.Associate(c, intId, extId.String())
			break
		}

		var ra annotation.Resource
		if a, ok := resource.Type().Annotations(c).Get(annotation.ResourceType); ok {
			ra = a.(annotation.Resource)
		} else {
			log.Debug("Using default Resource annotation", "resource", resource.Type().Name())
			ra = annotation.DefaultResource()
		}
		updateNeeded, recreateNeeded := ra.Changed(c, desiredState, result)

		if updateNeeded {
			if !recreateNeeded {
				// Update existing content. If an update method exists, call it. If not, then fall back
				// to delete + create
				if _, ok := crd.Member(`update`); ok {
					log.Debug("Update state", "extId", extId)
					result = px.AssertInstance(handlerDef.Label, resource.Type(), handler.Invoke(c, hn, `update`, extId, desiredState)).(px.PuppetObject)
					break
				}
			}

			// Rely on that deletion happens by means of LazyIdentity at end of run
			log.Debug("Remove external", "extId", extId)
			identity.RemoveExternal(c, extId.String())

			log.Debug("Create state", "intId", intId)
			rt := handler.Invoke(c, hn, `create`, desiredState)
			rl := rt.(px.List)
			result = px.AssertInstance(handlerDef.Label, resource.Type(), rl.At(0)).(px.PuppetObject)
			extId = rl.At(1)
			log.Debug("Associate state", "intId", intId, "extId", extId)
			identity.Associate(c, intId, extId.String())
		}
	default:
		panic(px.Error(wf.IllegalOperation, issue.H{`operation`: op}))
	}

	switch op {
	case wf.Read, wf.Upsert:
		returns := resource.Returns()
		entries := make([]*types.HashEntry, len(returns))
		for i, o := range returns {
			entries[i] = getValue(o, resource, result)
		}
		return types.WrapHash(entries)
	}
	return px.EmptyMap
}

func getValue(p serviceapi.Parameter, r wfe.Resource, o px.PuppetObject) *types.HashEntry {
	n := p.Name()
	a := n
	v := p.Alias()
	if v != `` {
		vs := strings.Split(v, `,`)
		if len(vs) > 1 {
			// Build hash from multiple attributes
			entries := make([]*types.HashEntry, len(vs))
			for i, a := range vs {
				if v, ok := o.Get(a); ok {
					entries[i] = types.WrapHashEntry2(a, v)
				} else {
					panic(px.Error(wfe.NoSuchAttribute, issue.H{`step`: r, `name`: a}))
				}
			}
			return types.WrapHashEntry2(n, types.WrapHash(entries))
		}
		a = v
	}
	if v, ok := o.Get(a); ok {
		return types.WrapHashEntry2(n, v)
	}
	panic(px.Error(wfe.NoSuchAttribute, issue.H{`step`: r, `name`: a}))
}
