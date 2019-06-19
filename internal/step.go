package internal

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
	"github.com/lyraproj/wfe/wfe"
)

type step struct {
	serviceId  px.TypedName
	name       string
	origin     string
	when       wf.Condition
	parameters []serviceapi.Parameter
	returns    []serviceapi.Parameter
	index      int
}

func CreateStep(c px.Context, def serviceapi.Definition) wfe.Step {
	hclog.Default().Debug(`creating step`, `style`, wfe.GetStringProperty(def, `style`))

	switch wfe.GetStringProperty(def, `style`) {
	case `stateHandler`:
		return newStateHandler(def)
	case `iterator`:
		return newIterator(c, def)
	case `resource`:
		return newResource(c, def)
	case `workflow`:
		return newWorkflow(c, def)
	case `action`:
		return newAction(def)
	case `call`:
		return newCall(c, def)
	}
	return nil
}

func (a *step) GetService(c px.Context) serviceapi.Service {
	return wfe.GetService(c, a.serviceId)
}

func (a *step) Origin() string {
	return a.origin
}

func (a *step) ServiceId() px.TypedName {
	return a.serviceId
}

func StepLabel(a wfe.Step) string {
	return fmt.Sprintf(`%s '%s'`, a.Style(), a.Name())
}

func StepId(a wfe.Step) string {
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

func (a *step) When() wf.Condition {
	return a.when
}

func (a *step) Name() string {
	return a.name
}

func (a *step) Parameters() []serviceapi.Parameter {
	return a.parameters
}

func (a *step) Returns() []serviceapi.Parameter {
	return a.returns
}

func (a *step) initStep(def serviceapi.Definition) {
	a.index = -1
	a.serviceId = def.ServiceId()
	a.name = def.Identifier().Name()
	props := def.Properties()
	if o, ok := props.Get4(`origin`); ok {
		a.origin = o.String()
	}
	a.parameters = getParameters(`parameters`, props)
	a.returns = getParameters(`returns`, props)
	if wh, ok := props.Get4(`when`); ok {
		a.when = wh.(wf.Condition)
	} else {
		a.when = wf.Always
	}
}

func getParameters(key string, props px.OrderedMap) []serviceapi.Parameter {
	if parameters, ok := props.Get4(key); ok {
		ia := parameters.(px.List)
		is := make([]serviceapi.Parameter, ia.Len())
		ia.EachWithIndex(func(iv px.Value, idx int) { is[idx] = iv.(serviceapi.Parameter) })
		return is
	}
	return []serviceapi.Parameter{}
}

func (a *step) IdParams() url.Values {
	if a.index >= 0 {
		return url.Values{`index`: {strconv.Itoa(a.index)}}
	}
	return url.Values{}
}

func resolveParameters(ctx px.Context, a wfe.Step, parameters px.OrderedMap) px.OrderedMap {
	ps := a.Parameters()
	ne := make([]*types.HashEntry, len(ps))
	for i, p := range ps {
		v, ok := parameters.Get4(p.Name())
		if !ok {
			v = p.Value()
			if v == nil {
				panic(px.Error(wfe.ParameterUnresolved, issue.H{`step`: a, `parameter`: p.Name()}))
			}
		}
		ne[i] = types.WrapHashEntry2(p.Name(), v)
	}
	return types.ResolveDeferred(ctx, types.WrapHash(ne), parameters).(px.OrderedMap)
}

// setIndex must only be called after a direct cloning operation on the instance, i.e. from WithIndex()
func (a *step) setIndex(index int) {
	a.index = index
}
