package puppet

import (
	"github.com/puppetlabs/go-evaluator/errors"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/wfe"
	"io"
)

type action struct {
	name string
	functions map[string]eval.InvocableValue
}

func NewActionBlock(name string, functions map[string]eval.InvocableValue) eval.InvocableValue {
	return &action{name, functions}
}

func (a *action) String() string {
	return a.name
}

func (a *action) Equals(other interface{}, guard eval.Guard) bool {
  return a == other
}

func (a *action) ToString(bld io.Writer, format eval.FormatContext, g eval.RDetect) {
	io.WriteString(bld, a.name)
}

func (a *action) Type() eval.PType {
	return a.functions[`read`].Type()
}

func (a *action) Call(c eval.Context, block eval.Lambda, args ...eval.PValue) (v eval.PValue) {
	if block != nil {
		panic(errors.NewArgumentsError(a.name, `blocks are not supported`))
	}

	ctx := wfe.ActivityContext(c)
	op := ctx.Get5(`operation`, types.WrapString(`read`)).String()
	switch op {
	case `create`:
		if f, ok := a.functions[op]; ok {
			return f.Call(c, nil, args...)
		}
		// No create, fall back to read
		if f, ok := a.functions[`read`]; ok {
			return f.Call(c, nil, args...)
		}
	case `read`:
		if f, ok := a.functions[op]; ok {
			return f.Call(c, nil, args...)
		}
	case `update`:
		if f, ok := a.functions[op]; ok {
			return f.Call(c, nil, args...)
		}
		// No update was present. If both create an delete exists, then use that instead
		if cf, ok := a.functions[`create`]; ok {
			if df, ok := a.functions[`delete`]; ok {
				df.Call(c, block, args...)
				return cf.Call(c, nil, args...)
			}
		}
		// Fall back to read
		if f, ok := a.functions[`read`]; ok {
			return f.Call(c, nil, args...)
		}
	case `delete`:
		if f, ok := a.functions[op]; ok {
			return f.Call(c, nil, args...)
		}
	}
	return
}
