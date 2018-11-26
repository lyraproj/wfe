package golang

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/impl"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-issues/issue"
	"reflect"
)

func MakeParams(ctx eval.Context, name string, v interface{}) []eval.Parameter {
	if v == nil {
		return []eval.Parameter{}
	}
	switch v.(type) {
	case []eval.Parameter:
		return v.([]eval.Parameter)
	default:
		return ParamsFromStruct(ctx, name, reflect.TypeOf(v))
	}
}

func ParamsFromStruct(ctx eval.Context, name string, ptr reflect.Type) []eval.Parameter {
	if ptr.Kind() == reflect.Ptr {
		s := ptr.Elem()
		if s.Kind() == reflect.Struct {
			outCount := s.NumField()
			params := make([]eval.Parameter, outCount)
			r := ctx.Reflector()
			for i := 0; i < outCount; i++ {
				fld := s.Field(i)
				name, decl := r.ReflectFieldTags(&fld)
				params[i] = impl.NewParameter(name, decl.Get5(`type`, types.DefaultAnyType()).(eval.Type), decl.Get5(`value`, nil), false)
			}
			return params
		}
	}
	panic(eval.Error(WF_NOT_STRUCTPTR, issue.H{`name`: name, `type`: ptr.String()}))
}
