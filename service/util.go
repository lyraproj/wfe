package service

import (
	"fmt"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wfapi"
	"github.com/lyraproj/wfe/api"
	"strings"
)

const ActivityContextKey = `activity::context`

func ActivityContext(c eval.Context) eval.OrderedMap {
	if ac, ok := c.Scope().Get(ActivityContextKey); ok {
		return eval.AssertInstance(`invalid activity context`, types.DefaultHashType(), ac).(eval.OrderedMap)
	}
	panic(eval.Error(api.WF_NO_ACTIVITY_CONTEXT, issue.NO_ARGS))
}

func GetOperation(ac eval.OrderedMap) wfapi.Operation {
	if op, ok := ac.Get4(`operation`); ok {
		return wfapi.Operation(op.(*types.IntegerValue).Int())
	}
	return wfapi.Read
}

func GetService(c eval.Context, serviceId eval.TypedName) serviceapi.Service {
	if serviceId.Namespace() == eval.NsService {
		if sm, ok := eval.Load(c, serviceId); ok {
			return sm.(serviceapi.Service)
		}
	}
	panic(eval.Error(api.WF_UNABLE_TO_LOAD_REQUIRED, issue.H{`namespace`: string(eval.NsService), `name`: serviceId.String()}))
}

func GetDefinition(c eval.Context, definitionId eval.TypedName) serviceapi.Definition {
	if definitionId.Namespace() == eval.NsDefinition {
		if sm, ok := eval.Load(c, definitionId); ok {
			return sm.(serviceapi.Definition)
		}
	}
	panic(eval.Error(api.WF_UNABLE_TO_LOAD_REQUIRED, issue.H{`namespace`: string(eval.NsDefinition), `name`: definitionId.String()}))
}

func GetHandler(c eval.Context, handlerId eval.TypedName) serviceapi.Definition {
	if handlerId.Namespace() == eval.NsHandler {
		if sm, ok := eval.Load(c, handlerId); ok {
			return sm.(serviceapi.Definition)
		}
	}
	panic(eval.Error(api.WF_UNABLE_TO_LOAD_REQUIRED, issue.H{`namespace`: string(eval.NsHandler), `name`: handlerId.String()}))
}

func GetStringProperty(def serviceapi.Definition, key string) string {
	return GetProperty(def, key, types.DefaultStringType()).String()
}

func GetProperty(def serviceapi.Definition, key string, typ eval.Type) eval.Value {
	if prop, ok := def.Properties().Get4(key); ok {
		return eval.AssertInstance(func() string {
			return fmt.Sprintf(`%s %s, property %s`, def.ServiceId(), def.Identifier(), key)
		}, typ, prop)
	}
	panic(eval.Error(api.WF_MISSING_REQUIRED_PROPERTY, issue.H{`service`: def.ServiceId(), `definition`: def.Identifier(), `key`: key}))
}

func GetOptionalProperty(def serviceapi.Definition, key string, typ eval.Type) (eval.Value, bool) {
	if prop, ok := def.Properties().Get4(key); ok {
		return eval.AssertInstance(func() string {
			return fmt.Sprintf(`%s %s, property %s`, def.ServiceId(), def.Identifier(), key)
		}, typ, prop), true
	}
	return nil, false
}

func LeafName(name string) string {
	names := strings.Split(name, `::`)
	return names[len(names)-1]
}
