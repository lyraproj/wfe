package wfe

import (
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-servicesdk/condition"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"github.com/puppetlabs/go-servicesdk/wfapi"
	"net/url"
	"strings"
)

const ActivityContextKey = `activity::context`

type Activity struct {
	serviceId eval.TypedName
	style  string
	name   string
	when   wfapi.Condition
	input  []eval.Parameter
	output []eval.Parameter
}

func CreateActivity(def serviceapi.Definition) api.Activity {
	switch GetStringProperty(def, `style`) {
	case `action`:
		return Action(def)
	case `iterator`:
		return Iterator(def)
	case `resource`:
		return Resource(def)
	case `workflow`:
		return Workflow(def)
	case `stateless`:
		return Stateless(def)
	}
	return nil
}

func GetService(c eval.Context, serviceId eval.TypedName) serviceapi.Service {
	if serviceId.Namespace() == eval.NsService {
		if sm, ok := eval.Load(c, serviceId); ok {
			return sm.(serviceapi.Service)
		}
	}
	panic(eval.Error(WF_UNABLE_TO_LOAD_REQUIRED, issue.H{`namespace`: string(eval.NsService), `name`: serviceId.String()}))
}

func GetDefinition(c eval.Context, definitionId eval.TypedName) serviceapi.Definition {
	if definitionId.Namespace() == eval.NsDefinition {
		if sm, ok := eval.Load(c, definitionId); ok {
			return sm.(serviceapi.Definition)
		}
	}
	panic(eval.Error(WF_UNABLE_TO_LOAD_REQUIRED, issue.H{`namespace`: string(eval.NsDefinition), `name`: definitionId.String()}))
}

func GetHandler(c eval.Context, handlerId eval.TypedName) serviceapi.Definition {
	if handlerId.Namespace() == eval.NsHandler {
		if sm, ok := eval.Load(c, handlerId); ok {
			return sm.(serviceapi.Definition)
		}
	}
	panic(eval.Error(WF_UNABLE_TO_LOAD_REQUIRED, issue.H{`namespace`: string(eval.NsHandler), `name`: handlerId.String()}))
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
	panic(eval.Error(WF_MISSING_REQUIRED_PROPERTY, issue.H{`service`: def.ServiceId(), `definition`: def.Identifier(), `key`: key}))
}

func GetOptionalProperty(def serviceapi.Definition, key string, typ eval.Type) (eval.Value, bool) {
	if prop, ok := def.Properties().Get4(key); ok {
		return eval.AssertInstance(func() string {
			return fmt.Sprintf(`%s %s, property %s`, def.ServiceId(), def.Identifier(), key)
		}, typ, prop), true
	}
	return nil, false
}

func ActivityContext(c eval.Context) eval.OrderedMap {
	if ac, ok := c.Scope().Get(ActivityContextKey); ok {
		return eval.AssertInstance(`invalid activity context`, types.DefaultHashType(), ac).(eval.OrderedMap)
	}
	panic(eval.Error(api.WF_NO_ACTIVITY_CONTEXT, issue.NO_ARGS))
}

func LeafName(activity api.Activity) string {
	names := strings.Split(activity.Name(), `::`)
	return names[len(names)-1]
}

func (a *Activity) GetService(c eval.Context) serviceapi.Service {
	return GetService(c, a.serviceId)
}

func (a *Activity) Style() string {
	return `activity`
}

func ActivityLabel(a api.Activity) string {
	return fmt.Sprintf(`%s '%s'`, a.Style(), a.Name())
}

func (a *Activity) When() wfapi.Condition {
	return a.when
}

func (a *Activity) Name() string {
	return a.name
}

func (a *Activity) Input() []eval.Parameter {
	return a.input
}

func (a *Activity) Output() []eval.Parameter {
	return a.output
}

func (a *Activity) Init(def serviceapi.Definition) {
	a.serviceId = def.ServiceId()
	a.name = def.Identifier().Name()
	props := def.Properties()
	a.input = getParameters(`input`, props)
	a.output = getParameters(`output`, props)
	if wh, ok := props.Get4(`when`); ok {
		a.when = wh.(wfapi.Condition)
	} else {
		a.when = condition.Always
	}
}

func getParameters(key string, props eval.OrderedMap) []eval.Parameter {
	if input, ok := props.Get4(key); ok {
		ia := input.(eval.List)
		is := make([]eval.Parameter, ia.Len())
		ia.EachWithIndex(func(iv eval.Value, idx int) { is[idx] = iv.(eval.Parameter) })
		return is
	}
	return []eval.Parameter{}
}

func (a *Activity) Identifier() string {
	return `genesis://puppet.com/` + a.Style() + `/` + url.PathEscape(a.name)
}

func ResolveInput(ctx eval.Context, a api.Activity, input eval.OrderedMap, p eval.Parameter) eval.Value {
	if !p.HasValue() {
		if v, ok := input.Get4(p.Name()); ok {
			return v
		}
		panic(eval.Error(WF_PARAMETER_UNRESOLVED, issue.H{`activity`: a, `parameter`: p.Name()}))
	}
	return types.ResolveDeferred(ctx, p.Value())
}
