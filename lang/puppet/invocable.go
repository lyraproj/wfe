package puppet

import (
	"github.com/puppetlabs/go-evaluator/errors"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-parser/parser"
	"io"
)

type invocable struct {
	name       string
	parameters []eval.Parameter
	signature  *types.CallableType
	body       parser.Expression
}

func NewInvocableBlock(name string, parameters []eval.Parameter, signature *types.CallableType, body parser.Expression) eval.InvocableValue {
	return &invocable{name, parameters, signature, body}
}

func (i *invocable) String() string {
	return i.name
}

func (i *invocable) Equals(other interface{}, guard eval.Guard) bool {
	return i == other
}

func (i *invocable) ToString(bld io.Writer, format eval.FormatContext, g eval.RDetect) {
	io.WriteString(bld, i.name)
}

func (i *invocable) Type() eval.Type {
	return i.signature
}

func (i *invocable) Call(c eval.Context, block eval.Lambda, args ...eval.Value) (v eval.Value) {
	if block != nil {
		panic(errors.NewArgumentsError(i.name, `nested lambdas are not supported`))
	}

	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case *errors.NextIteration:
				v = err.(*errors.NextIteration).Value()
			case *errors.Return:
				v = err.(*errors.Return).Value()
			default:
				panic(err)
			}
		}
	}()
	v = impl.CallBlock(c, i.name, i.parameters, i.signature, i.body, args)
	return
}
