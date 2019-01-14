package service

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/annotation"
	"github.com/lyraproj/servicesdk/wfapi"
	"github.com/lyraproj/wfe/api"
)

func ApplyState(c eval.Context, resource api.Resource, input eval.OrderedMap) eval.OrderedMap {
	ac := ActivityContext(c)
	op := GetOperation(ac)

	handlerDef := GetHandler(c, resource.HandlerId())
	crd := GetProperty(handlerDef, `interface`, types.NewTypeType(types.DefaultObjectType())).(eval.ObjectType)
	identity := getIdentity(c)
	handler := GetService(c, handlerDef.ServiceId())

	intId := types.WrapString(resource.Identifier())
	extId := resource.ExtId()
	explicitExtId := extId != nil
	if !explicitExtId {
		// external id must exist in order to do a read or delete
		extId = identity.getExternal(c, intId, op == wfapi.Read || op == wfapi.Delete)
	}

	var result eval.PuppetObject
	hn := handlerDef.Identifier().Name()
	switch op {
	case wfapi.Read:
		if extId == nil {
			return eval.EMPTY_MAP
		}
		result = eval.AssertInstance(handlerDef.Label, resource.Type(), handler.Invoke(c, hn, `read`, extId)).(eval.PuppetObject)

	case wfapi.Upsert:
		if explicitExtId {
			// An explicit externalId is for resources not managed by us. Only possible action
			// here is a read
			result = handler.Invoke(c, hn, `read`, extId).(eval.PuppetObject)
			break
		}

		desiredState := GetService(c, resource.ServiceId()).State(c, resource.Name(), input)
		if extId == nil {
			// Nothing exists yet. Create a new instance
			rt := handler.Invoke(c, hn, `create`, desiredState).(eval.List)
			result = eval.AssertInstance(handlerDef.Label, resource.Type(), rt.At(0)).(eval.PuppetObject)
			extId = rt.At(1)
			identity.associate(c, intId, extId)
			break
		}

		// Read current state and check if an update is needed
		updateNeeded := false
		recreateNeeded := false
		currentState := eval.AssertInstance(handlerDef.Label, resource.Type(), handler.Invoke(c, hn, `read`, extId)).(eval.PuppetObject)

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
					result = eval.AssertInstance(handlerDef.Label, resource.Type(), handler.Invoke(c, hn, `update`, extId, desiredState)).(eval.PuppetObject)
					break
				}
			}

			handler.Invoke(c, hn, `delete`, extId)
			identity.removeExternal(c, extId)

			rt := handler.Invoke(c, hn, `create`, desiredState)
			rl := rt.(eval.List)
			result = eval.AssertInstance(handlerDef.Label, resource.Type(), rl.At(0)).(eval.PuppetObject)
			extId = rl.At(1)
			identity.associate(c, intId, extId)
		} else {
			result = currentState
		}
	default:
		panic(eval.Error(wfapi.WF_ILLEGAL_OPERATION, issue.H{`operation`: op}))
	}

	switch op {
	case wfapi.Read, wfapi.Upsert:
		output := resource.Output()
		entries := make([]*types.HashEntry, len(output))
		for i, o := range output {
			entries[i] = getValue(o, resource, result)
		}
		return types.WrapHash(entries)
	}
	return eval.EMPTY_MAP
}

func getValue(p eval.Parameter, r api.Resource, o eval.PuppetObject) *types.HashEntry {
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
					panic(eval.Error(api.WF_NO_SUCH_ATTRIBUTE, issue.H{`activity`: r, `name`: a}))
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
	panic(eval.Error(api.WF_NO_SUCH_ATTRIBUTE, issue.H{`activity`: r, `name`: a}))
}
