package internal

import (
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/wfe"
)

type identity struct {
	id        string
	invokable serviceapi.Invokable
}

func (i *identity) Associate(c px.Context, internalID, externalID string) {
	i.invokable.Invoke(c, i.id, `associate`, types.WrapString(internalID), types.WrapString(externalID))
}

func (i *identity) AddReference(c px.Context, internalID, otherID string) {
	i.invokable.Invoke(c, i.id, `addReference`, types.WrapString(internalID), types.WrapString(otherID))
}

func (i *identity) BumpEra(c px.Context) {
	i.invokable.Invoke(c, i.id, `bumpEra`)
}

func (i *identity) Garbage(c px.Context, prefix string) px.List {
	result := i.invokable.Invoke(c, i.id, `garbage`, types.WrapString(prefix))
	if l, ok := result.(px.List); ok {
		return l
	}
	return nil
}

func (i *identity) GetExternal(c px.Context, internalId string) (string, bool) {
	result := i.invokable.Invoke(c, i.id, `getExternal`, types.WrapString(internalId))
	if ra, ok := result.(*types.Array); ok && ra.Len() == 2 {
		return ra.At(0).String(), ra.At(1).(px.Boolean).Bool()
	}
	return ``, false
}

func (i *identity) GetInternal(c px.Context, externalId string) (string, bool) {
	result := i.invokable.Invoke(c, i.id, `getInternal`, types.WrapString(externalId))
	if ra, ok := result.(*types.Array); ok && ra.Len() == 2 {
		return ra.At(0).String(), ra.At(1).(px.Boolean).Bool()
	}
	return ``, false
}

func (i *identity) Sweep(c px.Context, prefix string) {
	i.invokable.Invoke(c, i.id, `sweep`, types.WrapString(prefix))
}

func (i *identity) PurgeExternal(c px.Context, externalID string) {
	i.invokable.Invoke(c, i.id, `purgeExternal`, types.WrapString(externalID))
}

func (i *identity) PurgeInternal(c px.Context, internalID string) {
	i.invokable.Invoke(c, i.id, `purgeInternal`, types.WrapString(internalID))
}

func (i *identity) PurgeReferences(c px.Context, prefix string) {
	i.invokable.Invoke(c, i.id, `purgeReferences`, types.WrapString(prefix))
}

func (i *identity) RemoveExternal(c px.Context, externalID string) {
	i.invokable.Invoke(c, i.id, `removeExternal`, types.WrapString(externalID))
}

func (i *identity) RemoveInternal(c px.Context, internalID string) {
	i.invokable.Invoke(c, i.id, `removeInternal`, types.WrapString(internalID))
}

func (i *identity) Search(c px.Context, prefix string) px.List {
	result := i.invokable.Invoke(c, i.id, `search`, types.WrapString(prefix))
	if l, ok := result.(px.List); ok {
		return l
	}
	return nil
}

var IdentityId = px.NewTypedName(px.NsDefinition, "Identity::Service")

func GetIdentity(c px.Context) serviceapi.Identity {
	d := wfe.GetDefinition(c, IdentityId)
	return &identity{d.Identifier().Name(), wfe.GetService(c, d.ServiceId())}
}
