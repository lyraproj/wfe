package wfe

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/servicesdk/serviceapi"
)

type identity struct {
	id string
	invokable serviceapi.Invokable
}

func wrapString(s string) eval.Value {
	if s == `` {
		return eval.UNDEF
	}
	return types.WrapString(s)
}

func (i *identity) associate(c eval.Context, internalID, externalID string) {
	i.invokable.Invoke(c, i.id, `associate`, wrapString(internalID), wrapString(externalID))
}

func (i *identity) exists(c eval.Context, internalId string) bool {
	result := i.invokable.Invoke(c, i.id, `get_external`, wrapString(internalId)).(eval.List)
	return result.At(1).(*types.BooleanValue).Bool()
}

func (i *identity) getExternal(c eval.Context, internalId string, required bool) (string, bool) {
	result := i.invokable.Invoke(c, i.id, `get_external`, wrapString(internalId))
	if id, ok := result.(*types.StringValue); ok && id.String() != `` {
		return id.String(), ok
	}
	if required {
		panic(eval.Error(WF_UNABLE_TO_DETERMINE_EXTERNAL_ID, issue.H{`id`: internalId}))
	}
	return ``, false
}

func (i *identity) getInternal(c eval.Context, externalID string) (string, bool) {
	result := i.invokable.Invoke(c, i.id, `get_internal`, wrapString(externalID))
	if id, ok := result.(*types.StringValue); ok && id.String() != `` {
		return id.String(), ok
	}
	return ``, false
}

func (i *identity) removeExternal(c eval.Context, externalID string) {
	i.invokable.Invoke(c, i.id, `remove_external`,wrapString(externalID))
}

func (i *identity) removeInternal(c eval.Context, internalID string) {
	i.invokable.Invoke(c, i.id, `remove_internal`, wrapString(internalID))
}

var IdentityId = eval.NewTypedName(eval.NsDefinition, serviceapi.IdentityName)

func getIdentity(c eval.Context) *identity {
	idef := GetDefinition(c, IdentityId)
	return &identity{idef.Identifier().Name(), GetService(c, idef.ServiceId())}
}
