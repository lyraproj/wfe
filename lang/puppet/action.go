package puppet

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/wfe"
	"github.com/puppetlabs/go-issues/issue"
	"io"
)

type crd struct {
	name       string
	parameters []eval.Parameter
	functions  map[api.Operation]eval.InvocableValue
}

func NewCRD(name string, parameters []eval.Parameter, functions map[api.Operation]eval.InvocableValue) api.CRD {
	if _, ok := functions[api.Update]; ok {
		return &crud{crd{name, parameters, functions}}
	}
	return &crd{name, parameters, functions}
}

func (a *crd) Create(ctx eval.Context, input eval.OrderedMap) (eval.OrderedMap, error) {
	if f, ok := a.functions[api.Create]; ok {
		return wfe.CallInvocable(ctx, f, a.parameters, input)
	}
	panic(eval.Error(wfe.WF_MISSING_REQUIRED_BLOCK, issue.H{`name`: a.name, `block`: `create`}))
}

func (a *crd) Read(ctx eval.Context, input eval.OrderedMap) (eval.OrderedMap, error) {
	// No create, fall back to read
	if f, ok := a.functions[api.Read]; ok {
		return wfe.CallInvocable(ctx, f, a.parameters, input)
	}
	panic(eval.Error(wfe.WF_MISSING_REQUIRED_BLOCK, issue.H{`name`: a.name, `block`: `read`}))
}

func (a *crd) Delete(ctx eval.Context, input eval.OrderedMap) (eval.OrderedMap, error) {
	// No create, fall back to read
	if f, ok := a.functions[api.Delete]; ok {
		return wfe.CallInvocable(ctx, f, a.parameters, input)
	}
	return eval.EMPTY_MAP, nil
}

func (a *crd) String() string {
	return a.name
}

func (a *crd) Equals(other interface{}, guard eval.Guard) bool {
	return a == other
}

func (a *crd) ToString(bld io.Writer, format eval.FormatContext, g eval.RDetect) {
	io.WriteString(bld, a.name)
}

func (a *crd) Type() eval.Type {
	return a.functions[api.Read].Type()
}

type crud struct {
	crd
}

func NewCRUD(name string, parameters []eval.Parameter, functions map[api.Operation]eval.InvocableValue) api.CRUD {
	return &crud{crd{name, parameters, functions}}
}

func (c *crud) Update(ctx eval.Context, input eval.OrderedMap) (eval.OrderedMap, error) {
	if f, ok := c.functions[api.Update]; ok {
		return wfe.CallInvocable(ctx, f, c.parameters, input)
	}
	panic(eval.Error(wfe.WF_MISSING_REQUIRED_BLOCK, issue.H{`name`: c.name, `block`: `update`}))
}
