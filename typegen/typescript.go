package typegen

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"bytes"
	"strings"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-evaluator/utils"
	"fmt"
)

type tsGenerator struct {
	ctx eval.Context
	excludes []string
}

func NewTsGenerator(ctx eval.Context, excludes []string) Generator {
	return &tsGenerator{ctx, excludes}
}

func (g *tsGenerator) GenerateTypes(ts eval.TypeSet, ns []string, indent int, bld *bytes.Buffer) {
	ns = relativeNs(ns, ts.Name())
	for _, n := range ns {
		newLine(indent, bld)
		bld.WriteString(`export namespace `)
		bld.WriteString(n)
		bld.WriteString(` {`)
		indent += 2
	}
	newLine(indent, bld)
	leafName := nsName(ns, ts.Name())
	bld.WriteString(`export namespace `)
	bld.WriteString(leafName)
	bld.WriteString(` {`)
	indent += 2
	ns = append(append(make([]string, 0, len(ns) + 1), ns...), leafName)
	ts.Types().EachValue(func(t eval.PValue) { g.GenerateType(t.(eval.PType), ns, indent, bld) })
	for i := len(ns); i > 0; i-- {
		indent -= 2
		newLine(indent, bld)
		bld.WriteByte('}')
	}
	bld.WriteByte('\n')
}

func (g *tsGenerator) GenerateType(t eval.PType, ns []string, indent int, bld *bytes.Buffer) {
	if ts, ok := t.(eval.TypeSet); ok {
		g.GenerateTypes(ts, ns, indent, bld)
		return
	}

	if pt, ok := t.(eval.ObjectType); ok {
		bld.WriteByte('\n')
		newLine(indent, bld)
		bld.WriteString(`export class `)
		bld.WriteString(nsName(ns, pt.Name()))
		if ppt, ok := pt.Parent().(eval.ObjectType); ok {
			bld.WriteString(` extends `)
			bld.WriteString(nsName(ns, ppt.Name()))
		}
		bld.WriteString(` {`)
		indent += 2
		ai := pt.AttributesInfo()
		allAttrs, thisAttrs, superAttrs := g.toTsAttrs(pt, ns, ai.Attributes())
		appendFields(thisAttrs, indent, bld)
		bld.WriteByte('\n')
		appendConstructor(allAttrs, thisAttrs, superAttrs, indent, bld)
		bld.WriteByte('\n')
		appendPValueGetter(len(superAttrs) > 0, thisAttrs, indent, bld)
		bld.WriteByte('\n')
		appendPTypeGetter(pt.Name(), indent, bld)
		indent -= 2
		newLine(indent, bld)
		bld.WriteByte('}')
	} else {
		appendTsType(ns, t, bld)
	}
}

func (g *tsGenerator) ToTsType(ns []string, pType eval.PType) string {
	bld := bytes.NewBufferString(``)
	appendTsType(ns, pType, bld)
	return bld.String()
}

type tsAttribute struct {
	name string
	typ string
	value string
}

func (g *tsGenerator) toTsAttrs(t eval.ObjectType, ns []string, attrs []eval.Attribute) (allAttrs, thisAttrs, superAttrs []*tsAttribute) {
	allAttrs = make([]*tsAttribute, len(attrs))
	superAttrs = make([]*tsAttribute, 0)
	thisAttrs = make([]*tsAttribute, 0)
	for i, attr := range attrs {
		tsAttr := &tsAttribute{name: issue.CamelToSnakeCase(attr.Name()), typ: g.ToTsType(ns, attr.Type())}
		if attr.HasValue() {
			tsAttr.value = toTsValue(attr.Value(g.ctx))
		}
		if attr.Container() == t {
			thisAttrs = append(thisAttrs, tsAttr)
		} else {
			superAttrs = append(superAttrs, tsAttr)
		}
		allAttrs[i] = tsAttr
	}
	return
}

func appendFields(thisAttrs []*tsAttribute, indent int, bld *bytes.Buffer) {
	for _, attr := range thisAttrs {
		newLine(indent, bld)
		bld.WriteString(`readonly `)
		bld.WriteString(attr.name)
		bld.WriteString(`: `)
		bld.WriteString(attr.typ)
		bld.WriteString(`;`)
	}
	return
}

func appendConstructor(allAttrs, thisAttrs, superAttrs []*tsAttribute, indent int, bld *bytes.Buffer) {
	newLine(indent, bld)
	bld.WriteString(`constructor(`)
	appendParameters(allAttrs, indent, bld)
	bld.WriteString(`) {`)
	indent += 2
	if len(superAttrs) > 0 {
		newLine(indent, bld)
		bld.WriteString(`super({`)
		for i, attr := range superAttrs {
			if i > 0 {
				bld.WriteString(`, `)
			}
			bld.WriteString(attr.name)
			bld.WriteString(`: `)
			bld.WriteString(attr.name)
		}
		bld.WriteString(`});`)
	}
	for _, attr := range thisAttrs {
		newLine(indent, bld)
		bld.WriteString(`this.`)
		bld.WriteString(attr.name)
		bld.WriteString(` = `)
		bld.WriteString(attr.name)
		bld.WriteByte(';')
	}
	indent -= 2
	newLine(indent, bld)
	bld.WriteByte('}')
}

func appendPValueGetter(hasSuper bool, thisAttrs []*tsAttribute, indent int, bld *bytes.Buffer) {
	newLine(indent, bld)
	bld.WriteString(`__pvalue() : {[s: string]: any} {`)
	indent += 2
	newLine(indent, bld)
	if hasSuper {
		bld.WriteString(`let ih = super.__pvalue();`)
	} else {
		bld.WriteString(`let ih: {[s: string]: any} = {};`)
	}
	for _, attr := range thisAttrs {
		newLine(indent, bld)
		if attr.value != `` {
			bld.WriteString(`if(this.`)
			bld.WriteString(attr.name)
			bld.WriteString(` !== `)
			bld.WriteString(attr.value)
			bld.WriteString(`)`)
			indent += 2
			newLine(indent, bld)
		}
		bld.WriteString(`ih['`)
		bld.WriteString(attr.name)
		bld.WriteString(`'] = this.`)
		bld.WriteString(attr.name)
		bld.WriteString(`;`)
		if attr.value != `` {
			indent -= 2
		}
	}
	newLine(indent, bld)
	bld.WriteString(`return ih;`)
	indent -= 2
	newLine(indent, bld)
	bld.WriteByte('}')
}

func appendPTypeGetter(name string, indent int, bld *bytes.Buffer) {
	newLine(indent, bld)
	bld.WriteString(`__ptype() : string {`)
	indent += 2
	newLine(indent, bld)
	bld.WriteString(`return '`)
	bld.WriteString(name)
	bld.WriteString(`';`)
	indent -= 2
	newLine(indent, bld)
	bld.WriteByte('}')
}

func appendParameters(params []*tsAttribute, indent int, bld *bytes.Buffer) {
	indent += 2
	bld.WriteString(`{`)
	indent += 2
	for _, attr := range params {
		newLine(indent, bld)
		bld.WriteString(attr.name)
		if attr.value != `` {
			bld.WriteString(` = `)
			bld.WriteString(attr.value)
		}
		bld.WriteString(`,`)
	}
	bld.Truncate(bld.Len() - 1) // Truncate last comma
	indent -= 2
	newLine(indent, bld)
	bld.WriteString(`}: {`)
	indent += 2

	for _, attr := range params {
		newLine(indent, bld)
		bld.WriteString(attr.name)
		if  attr.value != `` {
			bld.WriteByte('?')
		}
		bld.WriteString(`: `)
		bld.WriteString(attr.typ)
		bld.WriteByte(',')
	}

	bld.Truncate(bld.Len() - 1) // Truncate last comma
	indent -= 2
	newLine(indent, bld)
	bld.WriteString(`}`)
}

func toTsValue(value eval.PValue) string {
	bld := bytes.NewBufferString(``)
	appendTsValue(value, bld)
	return bld.String()
}

func appendTsValue(value eval.PValue, bld *bytes.Buffer) {
	switch value.(type) {
	case *types.UndefValue:
		bld.WriteString(`null`)
	case *types.StringValue:
		utils.PuppetQuote(bld, value.String())
	case *types.BooleanValue, *types.IntegerValue, *types.FloatValue:
		bld.WriteString(value.String())
	case *types.ArrayValue:
		bld.WriteByte('[')
		value.(*types.ArrayValue).EachWithIndex(func(e eval.PValue, i int) {
			if i > 0 {
				bld.WriteString(`, `)
			}
			appendTsValue(e, bld)
		})
		bld.WriteByte(']')
	case *types.HashValue:
		bld.WriteByte('{')
		value.(*types.HashValue).EachWithIndex(func(e eval.PValue, i int) {
			ev := e.(*types.HashEntry)
			if i > 0 {
				bld.WriteString(`, `)
			}
			utils.PuppetQuote(bld, ev.Key().String())
			bld.WriteString(`: `)
			appendTsValue(ev.Value(), bld)
		})
		bld.WriteByte('}')
	}
}

func appendTsType(ns []string, pType eval.PType, bld *bytes.Buffer) {
	switch pType.(type) {
	case *types.BooleanType:
		bld.WriteString(`boolean`)
	case *types.IntegerType, *types.FloatType:
		bld.WriteString(`number`)
	case *types.StringType:
		bld.WriteString(`string`)
	case *types.OptionalType:
		appendTsType(ns, pType.(*types.OptionalType).ContainedType(), bld)
		bld.WriteString(` | null`)
	case *types.ArrayType:
		bld.WriteString(`Array<`)
		appendTsType(ns, pType.(*types.ArrayType).Type(), bld)
		bld.WriteString(`>`)
	case *types.VariantType:
		for i, v := range pType.(*types.VariantType).Types() {
			if i > 0 {
				bld.WriteString(` | `)
			}
			appendTsType(ns, v, bld)
		}
	case *types.HashType:
		ht := pType.(*types.HashType)
		bld.WriteString(`{[s: `)
		appendTsType(ns, ht.KeyType(), bld)
		bld.WriteString(`]: `)
		appendTsType(ns, ht.ValueType(), bld)
		bld.WriteString(`}`)
	case *types.EnumType:
		for i, s := range pType.(*types.EnumType).Parameters() {
			if i > 0 {
				bld.WriteString(` | `)
			}
			appendTsValue(s, bld)
		}
	case *types.TypeAliasType:
		bld.WriteString(nsName(ns, pType.(*types.TypeAliasType).Name()))
	case eval.ObjectType:
		bld.WriteString(nsName(ns, pType.(eval.ObjectType).Name()))
	}
}

func newLine(indent int, bld *bytes.Buffer) {
	bld.WriteByte('\n')
	for n := 0; n < indent; n++ {
		bld.WriteByte(' ')
	}
}

func relativeNs(ns []string, name string) []string {
	parts := strings.Split(name, `::`)
	if len(parts) == 1 {
		return []string{}
	}
	if len(ns) == 0 || isParent(ns, parts) {
		return parts[len(ns):len(parts)-1]
	}
	panic(fmt.Errorf("cannot generate %s in namespace %s", name, ns))
}

func nsName(ns []string, name string) string {
	parts := strings.Split(name, `::`)
	if isParent(ns, parts) {
		return strings.Join(parts[len(ns):], `.`)
	}
	return strings.Join(parts, `.`)
}

func isParent(ns, n []string) bool {
	top := len(ns)
	if top < len(n) {
		for idx := 0; idx < top; idx++ {
			if n[idx] != ns[idx] {
				return false
			}
		}
		return true
	}
	return false
}
