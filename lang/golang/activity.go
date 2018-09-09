package golang

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/wfe"
	"github.com/puppetlabs/go-issues/issue"
	"reflect"
	"runtime"
)

var errorType = reflect.TypeOf([]error{}).Elem()

type goActivity struct {
	wfe.Activity
	function interface{}
}

// Add a Go function as an activity with the given name. The function can zero or one parameter and the optional
// parameter must be a pointer to a struct. It must return a pointer to a struct
//
// The "input" declaration for the activity is reflected from the the fields in the struct parameter and the
// "output" declaration is reflected from the fields in the returned struct.
func Activity(ctx eval.Context, name string, function interface{}) api.Activity {
	ar := reflect.TypeOf(function)
	if ar.Kind() != reflect.Func {
		_, file, line, _ := runtime.Caller(1)
		panic(issue.NewReported(WF_NOT_FUNCTION, issue.SEVERITY_ERROR, issue.H{`name`: name, `type`: ar.String()}, issue.NewLocation(file, line, 0)))
	}

	inc := ar.NumIn()
	if ar.IsVariadic() || inc > 1 {
		panic(badActionFunction(name, ar))
	}

	oc := ar.NumOut()
	if !(oc == 1 && ar.Out(0).AssignableTo(errorType) || oc == 2 && ar.Out(1).AssignableTo(errorType)) {
		panic(badActionFunction(name, ar))
	}

	input := eval.NoParameters
	if inc == 1 {
		input = reflectStruct(ctx, name, ar, ar.In(0))
	}

	output := eval.NoParameters
	if oc == 2 {
		output = reflectStruct(ctx, name, ar, ar.Out(0))
	}
	g := &goActivity{function: function}
	g.Init(name, input, output, nil)
	return g
}

func (g *goActivity) Label() string {
	return wfe.ActivityLabel(g)
}

func reflectStruct(ctx eval.Context, name string, funcType, s reflect.Type) []eval.Parameter {
	if s.Kind() != reflect.Ptr {
		panic(badActionFunction(name, funcType))
	}
	if s.Elem().Kind() != reflect.Struct {
		panic(badActionFunction(name, funcType))
	}
	return ParamsFromStruct(ctx, name, s)
}

func badActionFunction(name string, typ reflect.Type) error {
	_, file, line, _ := runtime.Caller(2)
	return issue.NewReported(WF_BAD_FUNCTION, issue.SEVERITY_ERROR, issue.H{`name`: name, `type`: typ.String()}, issue.NewLocation(file, line, 0))
}

func (g *goActivity) Run(ctx eval.Context, input eval.KeyedValue) eval.KeyedValue {
	fv := reflect.ValueOf(g.function)
	fvType := fv.Type()

	params := make([]reflect.Value, 0)
	if fvType.NumIn() > 0 {
		inType := fvType.In(0).Elem()
		in := reflect.New(inType).Elem()
		t := in.NumField()
		r := ctx.Reflector()
		for i := 0; i < t; i++ {
			pn := issue.CamelToSnakeCase(inType.Field(i).Name)
			arg := input.Get5(pn, eval.UNDEF)
			in.Field(i).Set(r.Reflect(arg))
		}
		params = append(params, in.Addr())
	}

	result := fv.Call(params)
	expCount := 1
	if len(g.Output()) > 0 {
		expCount++
	}
	rn := len(result)
	if rn != expCount {
		panic(ctx.Error(nil, WF_BAD_RETURN_COUNT, issue.H{`activity`: g, `expected_count`: expCount, `actual_count`: rn}))
	}

	if rn == 1 {
		if err := result[0].Interface(); err != nil {
			panic(err)
		}
		return eval.EMPTY_MAP
	}

	err := result[1].Interface()
	if err != nil {
		panic(err)
	}

	rs := result[0]
	rt := rs.Type()
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rs = rs.Elem()
	}
	if rt.Kind() != reflect.Struct {
		panic(eval.Error(WF_NOT_STRUCT, issue.H{`activity`: g, `type`: rt.String()}))
	}
	fc := rt.NumField()
	entries := make([]*types.HashEntry, fc)
	for i := 0; i < fc; i++ {
		ft := rt.Field(i)
		v := rs.Field(i)
		n := issue.CamelToSnakeCase(ft.Name)
		if v.IsValid() {
			entries[i] = types.WrapHashEntry2(n, eval.Wrap(ctx, v))
		} else {
			entries[i] = types.WrapHashEntry2(n, eval.UNDEF)
		}
	}
	return types.WrapHash(entries)
}
