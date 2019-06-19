package internal

import (
	"sync"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/wfe"
)

type commandType int

const (
	bumpEra = commandType(iota)
	addRef
	purgeReferences
	sweep

	identityKey = `Lyra::Deferred::IdentityService`
)

type command struct {
	t    commandType
	args []px.Value
}

func (cmd *command) apply(c px.Context, is serviceapi.Identity) {
	switch cmd.t {
	case bumpEra:
		is.BumpEra(c)
	case addRef:
		is.AddReference(c, cmd.args[0].String(), cmd.args[1].String())
	case purgeReferences:
		is.PurgeReferences(c, cmd.args[0].String())
	case sweep:
		is.Sweep(c, cmd.args[0].String())
	}
}

type lazyIdentity struct {
	deferredCommands []*command
	service          serviceapi.Identity
}

func (gc *lazyIdentity) AddReference(c px.Context, internalId, otherId string) {
	if gc.service == nil {
		gc.deferredCommands = append(gc.deferredCommands, &command{addRef, []px.Value{types.WrapString(internalId), types.WrapString(otherId)}})
	} else {
		gc.service.AddReference(c, internalId, otherId)
	}
}

func (gc *lazyIdentity) Associate(c px.Context, internalID, externalID string) {
	gc.getIdentity(c).Associate(c, internalID, externalID)
}

func (gc *lazyIdentity) BumpEra(c px.Context) {
	if gc.service == nil {
		gc.deferredCommands = append(gc.deferredCommands, &command{bumpEra, nil})
	} else {
		gc.service.BumpEra(c)
	}
}

func (gc *lazyIdentity) Garbage(ctx px.Context, internalIDPrefix string) px.List {
	if gc.service == nil {
		return px.EmptyArray
	}
	return gc.service.Garbage(ctx, internalIDPrefix)
}

func (gc *lazyIdentity) GetExternal(c px.Context, internalId string) (string, bool) {
	return gc.getIdentity(c).GetExternal(c, internalId)
}

func (gc *lazyIdentity) GetInternal(ctx px.Context, externalID string) (string, bool) {
	return gc.getIdentity(ctx).GetInternal(ctx, externalID)
}

func (gc *lazyIdentity) PurgeExternal(ctx px.Context, externalID string) {
	gc.getIdentity(ctx).PurgeExternal(ctx, externalID)
}

func (gc *lazyIdentity) PurgeInternal(ctx px.Context, internalID string) {
	gc.getIdentity(ctx).PurgeInternal(ctx, internalID)
}

func (gc *lazyIdentity) PurgeReferences(ctx px.Context, internalIDPrefix string) {
	if gc.service == nil {
		gc.deferredCommands = append(gc.deferredCommands, &command{purgeReferences, []px.Value{types.WrapString(internalIDPrefix)}})
	} else {
		gc.service.PurgeReferences(ctx, internalIDPrefix)
	}
}

func (gc *lazyIdentity) RemoveExternal(c px.Context, externalID string) {
	gc.getIdentity(c).RemoveExternal(c, externalID)
}

func (gc *lazyIdentity) RemoveInternal(ctx px.Context, internalID string) {
	gc.getIdentity(ctx).RemoveInternal(ctx, internalID)
}

func (gc *lazyIdentity) Search(ctx px.Context, internalIDPrefix string) px.List {
	return gc.getIdentity(ctx).Search(ctx, internalIDPrefix)
}

func (gc *lazyIdentity) Sweep(ctx px.Context, internalIDPrefix string) {
	if gc.service == nil {
		gc.deferredCommands = append(gc.deferredCommands, &command{sweep, []px.Value{types.WrapString(internalIDPrefix)}})
	} else {
		gc.service.Sweep(ctx, internalIDPrefix)
	}
}

var liLock sync.Mutex

func GetLazyIdentity(c px.Context) (li serviceapi.Identity) {
	liLock.Lock()
	if v, ok := c.Get(identityKey); ok {
		li = v.(serviceapi.Identity)
	} else {
		li = &lazyIdentity{}
		c.Set(identityKey, li)
	}
	liLock.Unlock()
	return
}

func (gc *lazyIdentity) getIdentity(c px.Context) serviceapi.Identity {
	if gc.service == nil {
		d := wfe.GetDefinition(c, IdentityId)
		gc.service = &identity{d.Identifier().Name(), wfe.GetService(c, d.ServiceId())}
		for _, cmd := range gc.deferredCommands {
			cmd.apply(c, gc.service)
		}
		gc.deferredCommands = nil
	}
	return gc.service
}
