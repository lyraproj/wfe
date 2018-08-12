package typegen

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"bytes"
	"strings"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-evaluator/utils"
)

func GenerateType(c eval.Context, t eval.PType, ns []string, indent int, bld *bytes.Buffer) {
	if ts, ok := t.(eval.TypeSet); ok {
		GenerateTypes(c, ts, ns, indent, bld)
		return
	}

	if pt, ok := t.(eval.ObjectType); ok {
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
		superAttrs := make([]eval.Attribute, 0)
		for _, attr := range ai.Attributes() {
			if attr.Container() == pt {
				newLine(indent, bld)
				bld.WriteString(`readonly `)
				bld.WriteString(issue.CamelToSnakeCase(attr.Name()))
				bld.WriteString(`: `)
				toTsType(ns, attr.Type(), bld)
				bld.WriteString(`;`)
			} else {
				superAttrs = append(superAttrs, attr)
			}
		}
		bld.WriteByte('\n')
		newLine(indent, bld)
		bld.WriteString(`constructor(`)
		indent += 2
		newLine(indent, bld)
		bld.WriteString(`{`)
		indent += 2
		for _, attr := range ai.Attributes() {
			newLine(indent, bld)
			bld.WriteString(issue.CamelToSnakeCase(attr.Name()))
			if attr.HasValue() {
				bld.WriteString(` = `)
				appendValue(attr.Value(c), bld)
			}
			bld.WriteString(`,`)
		}
		bld.Truncate(bld.Len() - 1) // Truncate last comma
		indent -= 2
		newLine(indent, bld)
		bld.WriteString(`}: {`)
		indent += 2

		for _, attr := range ai.Attributes() {
			newLine(indent, bld)
			bld.WriteString(issue.CamelToSnakeCase(attr.Name()))
			if attr.HasValue() {
				bld.WriteByte('?')
			}
			bld.WriteString(`: `)
			toTsType(ns, attr.Type(), bld)
			bld.WriteByte(',')
		}

		bld.Truncate(bld.Len() - 1) // Truncate last comma
		indent -= 2
		newLine(indent, bld)
		bld.WriteString(`}) {`)
		if len(superAttrs) > 0 {
			newLine(indent, bld)
			bld.WriteString(`super({`)
			for i, attr := range superAttrs {
				if i > 0 {
					bld.WriteString(`, `)
				}
				n := issue.CamelToSnakeCase(attr.Name())
				bld.WriteString(n)
				bld.WriteString(`: `)
				bld.WriteString(n)
			}
			bld.WriteString(`});`)
		}
		for _, attr := range ai.Attributes() {
			if attr.Container() == pt {
				newLine(indent, bld)
				bld.WriteString(`this.`)
				n := issue.CamelToSnakeCase(attr.Name())
				bld.WriteString(n)
				bld.WriteString(` = `)
				bld.WriteString(n)
				bld.WriteByte(';')
			}
		}
		indent -= 2
		newLine(indent, bld)
		bld.WriteByte('}')
		indent -= 2
		newLine(indent, bld)
		bld.WriteByte('}')
	}
}

func appendValue(value eval.PValue, bld *bytes.Buffer) {
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
			appendValue(e, bld)
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
			appendValue(ev.Value(), bld)
		})
		bld.WriteByte('}')
	}
}

func toTsType(ns []string, pType eval.PType, bld *bytes.Buffer) {
	switch pType.(type) {
	case *types.BooleanType:
		bld.WriteString(`boolean`)
	case *types.IntegerType, *types.FloatType:
		bld.WriteString(`number`)
	case *types.StringType:
		bld.WriteString(`string`)
	case *types.OptionalType:
		toTsType(ns, pType.(*types.OptionalType).ContainedType(), bld)
		bld.WriteString(` | null`)
	case *types.ArrayType:
		bld.WriteString(`Array<`)
		toTsType(ns, pType.(*types.ArrayType).Type(), bld)
		bld.WriteString(`>`)
	case *types.VariantType:
		for i, v := range pType.(*types.VariantType).Types() {
			if i > 0 {
				bld.WriteString(` | `)
			}
			toTsType(ns, v, bld)
		}
	case *types.HashType:
		ht := pType.(*types.HashType)
		bld.WriteString(`{[s: `)
		toTsType(ns, ht.KeyType(), bld)
		bld.WriteString(`]: `)
		toTsType(ns, ht.ValueType(), bld)
		bld.WriteString(`}`)
	case *types.TypeAliasType:
		bld.WriteString(nsName(ns, pType.(*types.TypeAliasType).Name()))
	case eval.ObjectType:
		bld.WriteString(nsName(ns, pType.(eval.ObjectType).Name()))
	}
}

func GenerateTypes(c eval.Context, ts eval.TypeSet, ns []string, indent int, bld *bytes.Buffer) {
	newLine(indent, bld)
	leafName := nsName(ns, ts.Name())
	bld.WriteString(`export namespace `)
	bld.WriteString(leafName)
	bld.WriteString(` {`)
	indent += 2
	ns = append(append(make([]string, 0, len(ns) + 1), ns...), leafName)
	ts.Types().EachValue(func(t eval.PValue) { GenerateType(c, t.(eval.PType), ns, indent, bld) })
	indent -= 2;
	newLine(indent, bld);
	bld.WriteString("}\n");
}

func newLine(indent int, bld *bytes.Buffer) {
	bld.WriteByte('\n')
	for n := 0; n < indent; n++ {
		bld.WriteByte(' ')
	}
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
