package wfe

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wf"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/wfe/service"
)

type Activity struct {
	serviceId px.TypedName
	name      string
	when      wf.Condition
	input     []px.Parameter
	output    []px.Parameter
	index     int
}

func CreateActivity(c px.Context, def serviceapi.Definition) api.Activity {
	hclog.Default().Debug(`creating activity`, `style`, service.GetStringProperty(def, `style`))

	switch service.GetStringProperty(def, `style`) {
	case `stateHandler`:
		return StateHandler(def)
	case `iterator`:
		return Iterator(c, def)
	case `resource`:
		return Resource(c, def)
	case `workflow`:
		return Workflow(c, def)
	case `action`:
		return Action(def)
	}
	return nil
}

func (a *Activity) GetService(c px.Context) serviceapi.Service {
	return service.GetService(c, a.serviceId)
}

func (a *Activity) ServiceId() px.TypedName {
	return a.serviceId
}

func ActivityLabel(a api.Activity) string {
	return fmt.Sprintf(`%s '%s'`, a.Style(), a.Name())
}

func ActivityId(a api.Activity) string {
	b := bytes.NewBufferString(`lyra://puppet.com`)
	for _, s := range strings.Split(a.Name(), `::`) {
		b.WriteByte('/')
		b.WriteString(url.PathEscape(s))
	}
	vs := a.IdParams()
	if len(vs) > 0 {
		b.WriteByte('?')
		b.WriteString(vs.Encode())
	}
	return b.String()
}

func (a *Activity) When() wf.Condition {
	return a.when
}

func (a *Activity) Name() string {
	return a.name
}

func (a *Activity) Input() []px.Parameter {
	return a.input
}

func (a *Activity) Output() []px.Parameter {
	return a.output
}

func (a *Activity) Init(def serviceapi.Definition) {
	a.index = -1
	a.serviceId = def.ServiceId()
	a.name = def.Identifier().Name()
	props := def.Properties()
	a.input = getParameters(`input`, props)
	a.output = getParameters(`output`, props)
	if wh, ok := props.Get4(`when`); ok {
		a.when = wh.(wf.Condition)
	} else {
		a.when = wf.Always
	}
}

func getParameters(key string, props px.OrderedMap) []px.Parameter {
	if input, ok := props.Get4(key); ok {
		ia := input.(px.List)
		is := make([]px.Parameter, ia.Len())
		ia.EachWithIndex(func(iv px.Value, idx int) { is[idx] = iv.(px.Parameter) })
		return is
	}
	return []px.Parameter{}
}

func (a *Activity) IdParams() url.Values {
	if a.index >= 0 {
		return url.Values{`index`: {strconv.Itoa(a.index)}}
	}
	return url.Values{}
}

func ResolveInput(ctx px.Context, a api.Activity, input px.OrderedMap, p px.Parameter) px.Value {
	if !p.HasValue() {
		if v, ok := input.Get4(p.Name()); ok {
			return v
		}
		panic(px.Error(ParameterUnresolved, issue.H{`activity`: a, `parameter`: p.Name()}))
	}
	return types.ResolveDeferred(ctx, p.Value(), input)
}

// setIndex must only be called after a direct cloning operation on the instance, i.e. from WithIndex()
func (a *Activity) setIndex(index int) {
	a.index = index
}
