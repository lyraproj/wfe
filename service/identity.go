package service

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/api"
)

type identity struct {
	id        string
	invokable serviceapi.Invokable
}

func (i *identity) associate(c eval.Context, internalID, externalID eval.Value) {
	i.invokable.Invoke(c, i.id, `associate`, internalID, externalID)
}

func (i *identity) exists(c eval.Context, internalId eval.Value) bool {
	result := i.invokable.Invoke(c, i.id, `get_external`, internalId).(eval.List)
	return result.At(1).(*types.BooleanValue).Bool()
}

func (i *identity) getExternal(c eval.Context, internalId eval.Value, required bool) eval.Value {
	result := i.invokable.Invoke(c, i.id, `get_external`, internalId)
	if id, ok := result.(*types.StringValue); ok && id.String() != `` {
		return id
	}
	if required {
		panic(eval.Error(api.WF_UNABLE_TO_DETERMINE_EXTERNAL_ID, issue.H{`id`: internalId}))
	}
	return nil
}

func (i *identity) getInternal(c eval.Context, externalID eval.Value) (eval.Value, bool) {
	result := i.invokable.Invoke(c, i.id, `get_internal`, externalID)
	if id, ok := result.(*types.StringValue); ok && id.String() != `` {
		return id, ok
	}
	return nil, false
}

func (i *identity) removeExternal(c eval.Context, externalID eval.Value) {
	i.invokable.Invoke(c, i.id, `remove_external`, externalID)
}

func (i *identity) removeInternal(c eval.Context, internalID eval.Value) {
	i.invokable.Invoke(c, i.id, `remove_internal`, internalID)
}

var IdentityId = eval.NewTypedName(eval.NsDefinition, serviceapi.IdentityName)

func getIdentity(c eval.Context) *identity {
	idef := GetDefinition(c, IdentityId)
	return &identity{idef.Identifier().Name(), GetService(c, idef.ServiceId())}
}
