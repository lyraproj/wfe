package api

import (
	"reflect"
	"github.com/puppetlabs/go-issues/issue"
	"runtime"
)

type Action interface {
	Name() string

	Input() []Parameter

	Output() []Parameter

	Call(Genesis, map[string]reflect.Value) map[string]reflect.Value
}

type action struct {
	name     string
	input []Parameter
	output []Parameter
}

type producerAction struct {
	action

	producer Producer
}

type Producer interface {
	Call(Genesis, Action, map[string]reflect.Value) map[string]reflect.Value
}

func NewAction(name string, producer Producer, input, output interface{}) Action {
	return &producerAction{
		action: action{
			name: name,
			input: MakeParams(`input`, input),
			output: MakeParams(`output`, output),
		},
		producer: producer,
	}
}

func MakeParams(name string, v interface{}) []Parameter {
	if v == nil {
		return []Parameter{}
	}
	switch v.(type) {
	case []Parameter:
		return v.([]Parameter)
	default:
		ptr := reflect.TypeOf(v)
		if ptr.Kind() == reflect.Ptr {
			s := ptr.Elem()
			if s.Kind() == reflect.Struct {
				outCount := s.NumField()
				params := make([]Parameter, outCount)
				for i := 0; i < outCount; i++ {
					fld := s.Field(i)

					// TODO: Use tags to denote lookup?
					params[i] = NewParameter(issue.CamelToSnakeCase(fld.Name), fld.Type.String(), nil)
				}
				return params
			}
		}

		_, file, line, _ := runtime.Caller(2)
		panic(issue.NewReported(GENESIS_ACTION_NOT_STRUCTPTR,
			issue.SEVERITY_ERROR, issue.H{`name`: name, `type`: ptr.String()}, issue.NewLocation(file, line, 0)))
	}
}

func (a *action) Input() []Parameter {
	return a.input
}

func (a *action) Output() []Parameter {
	return a.output
}

func (a *action) Name() string {
	return a.name
}

func (a *producerAction) Call(g Genesis, args map[string]reflect.Value) map[string]reflect.Value {
	return a.producer.Call(g, a, args)
}
