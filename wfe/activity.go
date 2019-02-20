package wfe

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/condition"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wfapi"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/wfe/service"
	"net/url"
	"strings"
)

type Activity struct {
	serviceId eval.TypedName
	style     string
	name      string
	when      wfapi.Condition
	input     []eval.Parameter
	output    []eval.Parameter
}

func CreateActivity(def serviceapi.Definition) api.Activity {
	hclog.Default().Debug(`creating activity`, `style`, service.GetStringProperty(def, `style`))

	switch service.GetStringProperty(def, `style`) {
	case `stateHandler`:
		return StateHandler(def)
	case `iterator`:
		return Iterator(def)
	case `resource`:
		return Resource(def)
	case `workflow`:
		return Workflow(def)
	case `action`:
		return Action(def)
	}
	return nil
}

func (a *Activity) GetService(c eval.Context) serviceapi.Service {
	return service.GetService(c, a.serviceId)
}

func (a *Activity) ServiceId() eval.TypedName {
	return a.serviceId
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
	b := bytes.NewBufferString(`lyra://puppet.com`)
	for _, s := range strings.Split(a.name, `::`) {
		b.WriteByte('/')
		b.WriteString(url.PathEscape(s))
	}
	return b.String()
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
