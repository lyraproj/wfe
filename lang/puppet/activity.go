package puppet

import (
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-evaluator/utils"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/wfe"
	"github.com/puppetlabs/go-fsm/wfe/condition"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-parser/parser"
	"io"
)

type pActivity struct {
	activity   api.Activity
	parent     *pActivity
	properties *types.HashValue
	signature  *types.CallableType
	expression *parser.ActivityExpression
}

func (a *pActivity) Identifier() string {
	return a.activity.Identifier()
}

func (a *pActivity) Input() []eval.Parameter {
	return a.activity.Input()
}

func (a *pActivity) Activities() []api.Activity {
	if wf, ok := a.activity.(api.Workflow); ok {
		return wf.Activities()
	}
	return []api.Activity{}
}

func (a *pActivity) Output() []eval.Parameter {
	return a.activity.Output()
}

func (a *pActivity) Label() string {
	return wfe.ActivityLabel(a.activity)
}

func (a *pActivity) Equals(other interface{}, guard eval.Guard) bool {
	return a == other
}

func (a *pActivity) When() api.Condition {
	return a.activity.When()
}

func (a *pActivity) ToString(bld io.Writer, format eval.FormatContext, g eval.RDetect) {
	io.WriteString(bld, string(a.expression.Style()))
	utils.WriteByte(bld, ' ')
	io.WriteString(bld, a.Name())
}

func (a *pActivity) PType() eval.Type {
	return a.signature
}

func (a *pActivity) Signature() eval.Signature {
	return a.signature
}

func (a *pActivity) String() string {
	return eval.ToString(a)
}

func (a *pActivity) Dispatchers() []eval.Lambda {
	return []eval.Lambda{a}
}

func (a *pActivity) Name() string {
	return a.expression.Name()
}

func (a *pActivity) Parameters() []eval.Parameter {
	return a.activity.Input()
}

func init() {
	impl.NewPuppetActivity = func(expression *parser.ActivityExpression) eval.Function {
		return &pActivity{expression: expression}
	}
}

func (a *pActivity) Call(c eval.Context, block eval.Lambda, args ...eval.Value) eval.Value {
	names := a.signature.ParameterNames()
	entries := make([]*types.HashEntry, len(args))
	for i, arg := range args {
		entries[i] = types.WrapHashEntry2(names[i], arg)
	}
	return a.CallNamed(c, block, types.WrapHash(entries))
}

func (a *pActivity) CallNamed(c eval.Context, block eval.Lambda, args eval.OrderedMap) eval.Value {
	return a.Run(c, args)
}

func (a *pActivity) Run(c eval.Context, args eval.OrderedMap) eval.OrderedMap {
	return a.activity.Run(c, args)
}

func (a *pActivity) Resolve(c eval.Context) {
	if a.activity != nil {
		panic(fmt.Sprintf(`Attempt to resolve already resolved %s %s`, string(a.expression.Style()), a.Name()))
	}

	if props := a.expression.Properties(); props != nil {
		v := eval.Evaluate(c, props)
		dh, ok := v.(*types.HashValue)
		if !ok {
			panic(eval.Error(WF_FIELD_TYPE_MISMATCH, issue.H{`field`: `properties`, `expected`: `Hash`, `actual`: v.PType()}))
		}
		a.properties = dh
	}

	input := a.extractParameters(a.properties, `input`, a.inferInput)
	output := a.extractParameters(a.properties, `output`, func() []eval.Parameter { return eval.NoParameters })
	name := a.expression.Name()
	elems := make([]*types.StructElement, len(output))
	for i, op := range output {
		elems[i] = types.NewStructElement2(op.Name(), op.PType())
	}
	signature := types.NewCallableType(impl.CreateTupleType(input), types.NewStructType(elems), nil)

	var activity api.Activity
	switch a.expression.Style() {
	case parser.ActivityStyleWorkflow:
		activity = wfe.NewWorkflow(name, input, output, a.getWhen(), a.getActivities(c)...)
	case parser.ActivityStyleAction:
		activity = wfe.Action(name, a.getCRD(c, name, input, signature), input, output, a.getWhen())
	case parser.ActivityStyleResource:
		extId, _ := a.getStringProperty(`external_id`)
		activity = wfe.Resource(c, name, a.getResourceType(c), a.getState(c), extId, input, output, a.getWhen())
	default:
		panic(eval.Error(api.WF_UNSUPPORTED_ACTIVITY_STYLE, issue.H{`style`: string(a.expression.Style())}))
	}

	iterator := a.possibleIterator(activity)
	if iterator != activity {
		// Iterator changes our signature
		output = iterator.Output()
		elems = make([]*types.StructElement, len(output))
		for i, op := range output {
			elems[i] = types.NewStructElement2(op.Name(), op.PType())
		}
		signature = types.NewCallableType(impl.CreateTupleType(input), types.NewStructType(elems), nil)
		activity = iterator
	}
	a.activity = activity
	a.signature = signature
}

func (a *pActivity) Style() string {
	return string(a.expression.Style())
}

func (a *pActivity) inferInput() []eval.Parameter {
	// TODO:
	return eval.NoParameters
}

func (a *pActivity) inferOutput() []eval.Parameter {
	// TODO:
	return eval.NoParameters
}

func noParamsFunc() []eval.Parameter {
	return eval.NoParameters
}

func (a *pActivity) possibleIterator(activity api.Activity) api.Activity {
	if a.properties == nil {
		return activity
	}

	v, ok := a.properties.Get4(`iteration`)
	if !ok {
		return activity
	}

	iteratorDef, ok := v.(*types.HashValue)
	if !ok {
		panic(eval.Error(WF_FIELD_TYPE_MISMATCH, issue.H{`field`: `iteration`, `expected`: `Hash`, `actual`: v.PType()}))
	}

	v = iteratorDef.Get5(`function`, eval.UNDEF)
	style, ok := v.(*types.StringValue)
	if !ok {
		panic(eval.Error(WF_FIELD_TYPE_MISMATCH, issue.H{`field`: `iteration.style`, `expected`: `String`, `actual`: v}))
	}
	over := a.extractParameters(iteratorDef, `params`, noParamsFunc)
	vars := a.extractParameters(iteratorDef, `vars`, noParamsFunc)
	v = iteratorDef.Get5(`name`, eval.UNDEF)
	name, ok := v.(*types.StringValue)
	if !ok {
		panic(eval.Error(WF_FIELD_TYPE_MISMATCH, issue.H{`field`: `iteration.name`, `expected`: `String`, `actual`: v}))
	}
	return wfe.Iterator(api.NewIterationStyle(style.String()), activity, name.String(), over, vars)
}

// Extract activities from a Workflow definition
func (a *pActivity) getActivities(c eval.Context) []api.Activity {
	de := a.expression.Definition()
	if de == nil {
		return []api.Activity{}
	}

	block, ok := de.(*parser.BlockExpression)
	if !ok {
		panic(eval.Error(WF_FIELD_TYPE_MISMATCH, issue.H{`field`: `definition`, `expected`: `CodeBlock`, `actual`: de}))
	}

	// Block should only contain activity expressions or something is wrong.
	stmts := block.Statements()
	acs := make([]api.Activity, len(stmts))
	for i, stmt := range stmts {
		if as, ok := stmt.(*parser.ActivityExpression); ok {
			ac := &pActivity{parent: a, expression: as}
			ac.Resolve(c)
			acs[i] = ac
		} else if fn, ok := stmt.(*parser.FunctionDefinition); ok {
			fn := impl.NewPuppetFunction(fn)
			fn.Resolve(c)
			acs[i] = wfe.Stateless(c, fn, nil)
		} else {
			panic(eval.Error(WF_NOT_ACTIVITY, issue.H{`actual`: stmt}))
		}
	}
	return acs
}

func (a *pActivity) getCRD(c eval.Context, name string, input []eval.Parameter, signature *types.CallableType) api.CRD {
	de := a.expression.Definition()
	if de == nil {
		panic(c.Error(a.expression, WF_NO_DEFINITION, issue.NO_ARGS))
	}

	block, ok := de.(*parser.BlockExpression)
	if !ok {
		panic(c.Error(de, WF_FIELD_TYPE_MISMATCH, issue.H{`field`: `definition`, `expected`: `CodeBlock`, `actual`: de}))
	}

	var fs map[api.Operation]eval.InvocableValue
	hasFunctions := false
	for _, e := range block.Statements() {
		if _, ok = e.(*parser.FunctionDefinition); ok {
			hasFunctions = true
			break
		}
	}
	if hasFunctions {
		// Block must only consist of functions the functions create, read, update, and delete.
		fs = make(map[api.Operation]eval.InvocableValue, len(block.Statements()))
		for _, e := range block.Statements() {
			if fd, ok := e.(*parser.FunctionDefinition); ok {
				switch fd.Name() {
				case `create`, `read`, `update`, `delete`:
					f := impl.NewPuppetFunction(fd)
					f.Resolve(c)
					fs[api.NewOperation(fd.Name())] = f
					continue
				}
			}
			panic(c.Error(e, WF_FIELD_TYPE_MISMATCH, issue.H{`field`: `definition`, `expected`: `Function create, read, update, or delete`, `actual`: fs}))
		}
	} else {
		fs = map[api.Operation]eval.InvocableValue{api.Read: NewInvocableBlock(name, a.Input(), signature, block)}
	}

	return NewCRD(name, input, fs)
}

func (a *pActivity) getWhen() api.Condition {
	if when, ok := a.getStringProperty(`when`); ok {
		return condition.Parse(when)
	}
	return nil
}

func (a *pActivity) extractParameters(props *types.HashValue, field string, dflt func() []eval.Parameter) []eval.Parameter {
	if props == nil {
		return dflt()
	}

	v, ok := props.Get4(field)
	if !ok {
		return dflt()
	}

	ia, ok := v.(*types.ArrayValue)
	if !ok {
		panic(eval.Error(WF_FIELD_TYPE_MISMATCH, issue.H{`field`: field, `expected`: `Array`, `actual`: v.PType()}))
	}

	params := make([]eval.Parameter, ia.Len())
	ia.EachWithIndex(func(v eval.Value, i int) {
		if p, ok := v.(eval.Parameter); ok {
			params[i] = p
		} else {
			panic(eval.Error(WF_ELEMENT_NOT_PARAMETER, issue.H{`type`: p.PType(), `field`: field}))
		}
	})
	return params
}

func (a *pActivity) getState(c eval.Context) eval.OrderedMap {
	de := a.expression.Definition()
	if de == nil {
		return eval.EMPTY_MAP
	}

	if hash, ok := de.(*parser.LiteralHash); ok {
		// Transform all variable references to Deferred expressions
		return eval.Evaluate(c, hash).(eval.OrderedMap)
	}
	panic(eval.Error(WF_FIELD_TYPE_MISMATCH, issue.H{`field`: `definition`, `expected`: `Hash`, `actual`: de}))
}

func (a *pActivity) getResourceType(c eval.Context) eval.ObjectType {
	n := a.Name()
	if a.properties != nil {
		if tv, ok := a.properties.Get4(`type`); ok {
			if t, ok := tv.(eval.ObjectType); ok {
				return t
			}
			if s, ok := tv.(*types.StringValue); ok {
				n = s.String()
			} else {
				panic(eval.Error(WF_FIELD_TYPE_MISMATCH, issue.H{`field`: `definition`, `expected`: `Variant[String,ObjectType]`, `actual`: tv}))
			}
		} else {
			ts := a.getTypespace()
			if ts != `` {
				n = ts + `::` + wfe.LeafName(a)
			}
		}
	}
	tn := eval.NewTypedName(eval.TYPE, n)
	if t, ok := eval.Load(c, tn); ok {
		if pt, ok := t.(eval.ObjectType); ok {
			return pt
		}
		panic(eval.Error(WF_FIELD_TYPE_MISMATCH, issue.H{`field`: `definition`, `expected`: `ObjectType`, `actual`: t}))
	}
	panic(eval.Error(eval.EVAL_UNRESOLVED_TYPE, issue.H{`typeString`: tn.Name()}))
}

func (a *pActivity) getTypespace() string {
	if ts, ok := a.getStringProperty(`typespace`); ok {
		return ts
	}
	if a.parent != nil {
		return a.parent.getTypespace()
	}
	return ``
}

func (a *pActivity) getStringProperty(field string) (string, bool) {
	if a.properties == nil {
		return ``, false
	}

	v, ok := a.properties.Get4(field)
	if !ok {
		return ``, false
	}

	if s, ok := v.(*types.StringValue); ok {
		return s.String(), true
	}
	panic(eval.Error(WF_FIELD_TYPE_MISMATCH, issue.H{`field`: field, `expected`: `String`, `actual`: v.PType()}))
}
