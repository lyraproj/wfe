package api

import (
	"context"
	"github.com/puppetlabs/go-issues/issue"
	"reflect"
	"runtime"
)

type goActionCall struct {
	function interface{}
}

var contextType = reflect.TypeOf([]context.Context{}).Elem()
var errorType = reflect.TypeOf([]error{}).Elem()
var noParams = make([]Parameter, 0)

func NewGoAction(name string, function interface{}) Action {
	ar := reflect.TypeOf(function)
	if ar.Kind() != reflect.Func {
		_, file, line, _ := runtime.Caller(1)
		panic(issue.NewReported(GENESIS_ACTION_NOT_FUNCTION, issue.SEVERITY_ERROR, issue.H{`name`: name, `type`: ar.String()}, issue.NewLocation(file, line, 0)))
	}

	inc := ar.NumIn()
	if ar.IsVariadic() || inc == 0 || inc > 2 || !ar.In(0).AssignableTo(contextType) {
		panic(badActionFunction(name, ar))
	}

	oc := ar.NumOut()
	if !(oc == 1 && ar.Out(0).AssignableTo(errorType) || oc == 2 && ar.Out(1).AssignableTo(errorType)) {
		panic(badActionFunction(name, ar))
	}

	input := noParams
	if inc == 2 {
		input = reflectStruct(name, ar, ar.In(1))
	}

	output := noParams
	if oc == 2 {
		output = reflectStruct(name, ar, ar.Out(0))
	}
	return NewAction(name, &goActionCall{function}, input, output)
}

func reflectStruct(name string, funcType, s reflect.Type) []Parameter {
	if s.Kind() != reflect.Ptr {
		panic(badActionFunction(name, funcType))
	}

	s = s.Elem()
	if s.Kind() != reflect.Struct {
		panic(badActionFunction(name, funcType))
	}
	outCount := s.NumField()
	params := make([]Parameter, outCount)
	for i := 0; i < outCount; i++ {
		fld := s.Field(i)

		// TODO: Use tags to denote lookup?
		params[i] = NewParameter(issue.CamelToSnakeCase(fld.Name), fld.Type.String(), nil)
	}
	return params
}

func badActionFunction(name string, typ reflect.Type) error {
	_, file, line, _ := runtime.Caller(2)
	return issue.NewReported(GENESIS_ACTION_BAD_FUNCTION, issue.SEVERITY_ERROR, issue.H{`name`: name, `type`: typ.String()}, issue.NewLocation(file, line, 0))
}

func (ga *goActionCall) Call(g Genesis, a Action, args map[string]reflect.Value) map[string]reflect.Value {
	fv := reflect.ValueOf(ga.function)
	fvType := fv.Type()

	params := make([]reflect.Value, 0, 2)
	params = append(params, reflect.ValueOf(g))
	if fvType.NumIn() > 1 {
		inType := fvType.In(1).Elem()
		in := reflect.New(inType).Elem()
		t := in.NumField()
		for i := 0; i < t; i++ {
			pn := inType.Field(i).Name
			in.Field(i).Set(args[issue.CamelToSnakeCase(pn)])
		}
		params = append(params, in.Addr())
	}
	result := fv.Call(params)
	expCount := 1
	if len(a.Output()) > 0 {
		expCount++
	}
	rn := len(result)
	if rn != expCount {
		panic(issue.NewReported(GENESIS_ACTION_BAD_RETURN_COUNT, issue.SEVERITY_ERROR, issue.H{`name`: a.Name(), `expected_count`: expCount, `actual_count`: rn}, nil))
	}

	if rn == 1 {
		if err := result[0].Interface(); err != nil {
			panic(err)
		}
		return map[string]reflect.Value{}
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
		panic(issue.NewReported(GENESIS_ACTION_NOT_STRUCT, issue.SEVERITY_ERROR, issue.H{`name`: a.Name(), `type`: rt.String()}, nil))
	}
	fc := rt.NumField()
	rm := make(map[string]reflect.Value, fc)
	for i := 0; i < fc; i++ {
		ft := rt.Field(i)
		v := rs.Field(i)
		n := issue.CamelToSnakeCase(ft.Name)
		if v.IsValid() {
			rm[n] = v
		} else {
			rm[n] = reflect.Zero(ft.Type)
		}
	}
	return rm
}
