package service

import (
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/annotation"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wf"
	"github.com/lyraproj/wfe/api"
)

type commandType int

const (
	startEra = commandType(iota)
	addRef

	gcKey = `Lyra::Deferred::IdentityService`
)

type command struct {
	t    commandType
	args []px.Value
}

func (cmd *command) apply(c px.Context, is *identity) {
	switch cmd.t {
	case startEra:
		is.bumpEra(c)
	case addRef:
		is.addReference(c, cmd.args[0], cmd.args[1])
	}
}

type LazyIdentity struct {
	deferredCommands []*command
	service          *identity
}

var liLock sync.Mutex

func GetLazyIdentity(c px.Context) (li *LazyIdentity) {
	liLock.Lock()
	if v, ok := c.Get(gcKey); ok {
		li = v.(*LazyIdentity)
	} else {
		li = &LazyIdentity{}
		c.Set(gcKey, li)
	}
	liLock.Unlock()
	return
}

func (gc *LazyIdentity) StartEra(c px.Context) {
	if gc.service == nil {
		gc.deferredCommands = append(gc.deferredCommands, &command{startEra, nil})
	} else {
		gc.service.bumpEra(c)
	}
}

// LazySweepAndGC calls SweepAndGC if there has been any activity detected since
// the instance was obtained. Bumping the era is not considered an activity
func (gc *LazyIdentity) LazySweepAndGC(c px.Context, prefix string) {
	if gc.service == nil {
		return
	}
	gc.SweepAndGC(c, prefix)
}

// SweepAndGC performs a sweep of the Identity store, retrieves all garbage, and
// then tells the handler for each garbage entry to delete the resource. The entry
// is then purged from the Identity store
func (gc *LazyIdentity) SweepAndGC(c px.Context, prefix string) {
	identity := gc.getIdentity(c)
	log := hclog.Default()
	log.Debug("LazyIdentity Sweep", "prefix", prefix)
	identity.sweep(c, prefix)
	log.Debug("LazyIdentity Collect garbage", "prefix", prefix)
	gl := identity.garbage(c, prefix)
	ng := gl.Len()
	log.Debug("LazyIdentity Collect garbage", "prefix", prefix, "count", ng)
	rs := make([]px.List, ng)

	// Store in reverse order
	ng--
	gl.EachWithIndex(func(t px.Value, i int) {
		rs[ng-i] = t.(px.List)
	})

	for _, l := range rs {
		uri := types.ParseURI(l.At(0).String())
		hid := uri.Query().Get(`hid`)
		if hid == `` {
			continue
		}
		handlerDef := GetHandler(c, px.NewTypedName(px.NsHandler, hid))
		handler := GetService(c, handlerDef.ServiceId())

		extId := l.At(1)
		log.Debug("LazyIdentity delete", "prefix", prefix, "intId", uri.String(), "extId", extId)
		handler.Invoke(c, handlerDef.Identifier().Name(), `delete`, extId)
		identity.purgeExternal(c, extId)
	}
	identity.purgeReferences(c, prefix)
}

func (gc *LazyIdentity) AddReference(c px.Context, internalId, otherId string) {
	iv := types.WrapString(internalId)
	ov := types.WrapString(otherId)
	if gc.service == nil {
		gc.deferredCommands = append(gc.deferredCommands, &command{addRef, []px.Value{iv, ov}})
	} else {
		gc.service.addReference(c, iv, ov)
	}
}

func (gc *LazyIdentity) getIdentity(c px.Context) *identity {
	if gc.service == nil {
		d := GetDefinition(c, IdentityId)
		gc.service = &identity{d.Identifier().Name(), GetService(c, d.ServiceId())}
		for _, cmd := range gc.deferredCommands {
			cmd.apply(c, gc.service)
		}
		gc.deferredCommands = nil
	}
	return gc.service
}

func (gc *LazyIdentity) readOrNotFound(c px.Context, handler serviceapi.Service, hn string, extId px.Value) px.Value {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(issue.Reported); ok {
				if e, ok = e.Cause().(issue.Reported); ok && e.Code() == service.NotFound {
					// Not found by remote. Purge the extId and return nil.
					hclog.Default().Debug("Removing obsolete extId from Identity service", "extId", extId)
					gc.getIdentity(c).purgeExternal(c, extId)
					return
				}
			}
			panic(r)
		}
	}()

	hclog.Default().Debug("Read state", "extId", extId)
	return handler.Invoke(c, hn, `read`, extId)
}

func (gc *LazyIdentity) getExternal(c px.Context, internalId px.Value, required bool) px.Value {
	return gc.getIdentity(c).getExternal(c, internalId, required)
}

func (gc *LazyIdentity) associate(c px.Context, internalID, externalID px.Value) {
	gc.getIdentity(c).associate(c, internalID, externalID)
}

func (gc *LazyIdentity) removeExternal(c px.Context, externalID px.Value) {
	gc.getIdentity(c).removeExternal(c, externalID)
}

func ApplyState(c px.Context, resource api.Resource, parameters px.OrderedMap) px.OrderedMap {
	ac := StepContext(c)
	op := GetOperation(ac)

	log := hclog.Default()
	handlerDef := GetHandler(c, resource.HandlerId())
	crd := GetProperty(handlerDef, `interface`, types.NewTypeType(types.DefaultObjectType())).(px.ObjectType)
	identity := GetLazyIdentity(c)
	handler := GetService(c, handlerDef.ServiceId())

	intId := types.WrapString(resource.Identifier())
	extId := resource.ExtId()
	explicitExtId := extId != nil
	if !explicitExtId {
		// external id must exist in order to do a read or delete
		extId = identity.getExternal(c, intId, op == wf.Read || op == wf.Delete)
		log.Debug("GetExternal", "intId", intId, "extId", extId)
	}

	var result px.PuppetObject
	hn := handlerDef.Identifier().Name()
	switch op {
	case wf.Read:
		if extId == nil {
			return px.EmptyMap
		}
		rt := identity.readOrNotFound(c, handler, hn, extId)
		if rt == nil {
			return px.EmptyMap
		}
		result = px.AssertInstance(handlerDef.Label, resource.Type(), rt).(px.PuppetObject)

	case wf.Upsert:
		if explicitExtId {
			// An explicit externalId is for resources not managed by us. Only possible action
			// here is a read
			rt := identity.readOrNotFound(c, handler, hn, extId)
			if rt == nil {
				// False positive from the Identity service
				return px.EmptyMap
			}
			result = px.AssertInstance(handlerDef.Label, resource.Type(), rt).(px.PuppetObject)
			break
		}

		desiredState := GetService(c, resource.ServiceId()).State(c, resource.Name(), parameters)
		if extId != nil {
			// Read current state and check if an update is needed
			rt := identity.readOrNotFound(c, handler, hn, extId)
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
			identity.associate(c, intId, extId)
			break
		}

		var updateNeeded, recreateNeeded bool
		if a, ok := resource.Type().Annotations(c).Get(annotation.ResourceType); ok {
			ra := a.(annotation.Resource)
			updateNeeded, recreateNeeded = ra.Changed(c, desiredState, result)
		} else {
			updateNeeded = !desiredState.Equals(result, nil)
			recreateNeeded = false
		}

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
			identity.removeExternal(c, extId)

			log.Debug("Create state", "intId", intId)
			rt := handler.Invoke(c, hn, `create`, desiredState)
			rl := rt.(px.List)
			result = px.AssertInstance(handlerDef.Label, resource.Type(), rl.At(0)).(px.PuppetObject)
			extId = rl.At(1)
			log.Debug("Associate state", "intId", intId, "extId", extId)
			identity.associate(c, intId, extId)
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

func getValue(p serviceapi.Parameter, r api.Resource, o px.PuppetObject) *types.HashEntry {
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
					panic(px.Error(api.NoSuchAttribute, issue.H{`step`: r, `name`: a}))
				}
			}
			return types.WrapHashEntry2(n, types.WrapHash(entries))
		}
		a = v
	}
	if v, ok := o.Get(a); ok {
		return types.WrapHashEntry2(n, v)
	}
	panic(px.Error(api.NoSuchAttribute, issue.H{`step`: r, `name`: a}))
}
