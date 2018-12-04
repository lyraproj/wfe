package wfe

import (
	"github.com/lyraproj/issue/issue"
)

const (
	WF_GRAPH_DOT_MARSHAL                 = `WF_GRAPH_DOT_MARSHAL`
	WF_ALREADY_DEFINED                   = `WF_ALREADY_DEFINED`
	WF_MULTIPLE_PRODUCERS_OF_VALUE       = `WF_MULTIPLE_PRODUCERS_OF_VALUE`
	WF_NO_PRODUCER_OF_VALUE              = `WF_NO_PRODUCER_OF_VALUE`
	WF_BAD_CONSUMES_COUNT                = `WF_BAD_CONSUMES_COUNT`
	WF_BAD_RETURN_TYPE                   = `WF_BAD_RETURN_TYPE`
	WF_ITERATION_ACTIVITY_WRONG_INPUT    = `WF_ITERATION_ACTIVITY_WRONG_INPUT`
	WF_ITERATION_ACTIVITY_WRONG_OUTPUT   = `WF_ITERATION_ACTIVITY_WRONG_OUTPUT`
	WF_ITERATION_PARAMETER_INVALID_COUNT = `WF_ITERATION_PARAMETER_INVALID_COUNT`
	WF_ITERATION_PARAMETER_WRONG_TYPE    = `WF_ITERATION_PARAMETER_WRONG_TYPE`
	WF_ITERATION_PARAMETER_UNRESOLVED    = `WF_ITERATION_PARAMETER_UNRESOLVED`
	WF_ITERATION_VARIABLE_INVALID_COUNT  = `WF_ITERATION_VARIABLE_INVALID_COUNT`
	WF_MISSING_REQUIRED_PROPERTY         = `WF_MISSING_REQUIRED_PROPERTY`
	WF_MULTIPLE_DISPATCHERS              = `WF_MULTIPLE_DISPATCHERS`
	WF_NO_SUCH_ATTRIBUTE                 = `WF_NO_SUCH_ATTRIBUTE`
	WF_OPERATION_DID_NOT_RETURN_STATE    = `WF_OPERATION_DID_NOT_RETURN_STATE`
	WF_PARAMETER_UNRESOLVED              = `WF_PARAMETER_UNRESOLVED`
	WF_TOO_MANY_GUARDS                   = `WF_TOO_MANY_GUARDS`
	WF_UNABLE_TO_REFLECT_TYPE            = `WF_UNABLE_TO_REFLECT_TYPE`
	WF_UNABLE_TO_LOAD_REQUIRED           = `WF_UNABLE_TO_LOAD_REQUIRED`
	WF_UNABLE_TO_DETERMINE_EXTERNAL_ID   = `WF_UNABLE_TO_DETERMINE_EXTERNAL_ID`
	WF_OUTPUT_NOT_STRUCT                 = `WF_OUTPUT_NOT_STRUCT`
	WF_MISSING_REQUIRED_BLOCK            = `WF_MISSING_REQUIRED_BLOCK`
)

func init() {
	issue.Hard(WF_GRAPH_DOT_MARSHAL, `error while marshalling graph to dot: %{detail}`)
	issue.Hard2(WF_ALREADY_DEFINED, `%{activity} is already defined`, issue.HF{`activity`: issue.Label})
	issue.Hard2(WF_MULTIPLE_PRODUCERS_OF_VALUE, `both %{activity1} and %{activity2} output the value '%{value}'`, issue.HF{`activity1`: issue.Label, `activity2`: issue.Label})
	issue.Hard2(WF_NO_PRODUCER_OF_VALUE, `%{activity} value '%{value}' is never produced`, issue.HF{`activity`: issue.Label})
	issue.Hard2(WF_OPERATION_DID_NOT_RETURN_STATE, `%{handler} %{op} did not return a state`, issue.HF{`handler`: issue.Label})
	issue.Hard(WF_UNABLE_TO_REFLECT_TYPE, `unable to convert type %{type} into a reflect.Type`)
	issue.Hard(WF_UNABLE_TO_LOAD_REQUIRED, `unable to load required %{namespace} '%{name}'`)
	issue.Hard(WF_UNABLE_TO_DETERMINE_EXTERNAL_ID, `unable to determine external ID for %{style} '%{id}'`)
	issue.Hard2(WF_BAD_CONSUMES_COUNT, `%{activity} expects %{expected} parameters, got %{actual}`, issue.HF{`activity`: issue.Label})

	issue.Hard2(WF_ITERATION_ACTIVITY_WRONG_INPUT, `%{iterator} input must consume output produced by the iterator`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_ACTIVITY_WRONG_OUTPUT, `%{iterator} output must consist of a 'key' and a 'value'`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_PARAMETER_INVALID_COUNT, `%{iterator} wrong number of parameters. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_PARAMETER_INVALID_COUNT, `%{iterator} wrong number of parameters. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_PARAMETER_WRONG_TYPE, `%{iterator} parameter %{parameter} if of wrong type. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_PARAMETER_UNRESOLVED, `%{iterator}, parameter %{parameter} cannot be resolved`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_VARIABLE_INVALID_COUNT, `%{iterator}, wrong number of variables. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})

	issue.Hard(WF_MISSING_REQUIRED_PROPERTY, `definition %{service} %{definition} is missing required property '%{key}'`)

	issue.Hard(WF_MULTIPLE_DISPATCHERS, `'%{name}' has more than one dispatcher`)
	issue.Hard2(WF_NO_SUCH_ATTRIBUTE, `%{activity} has no attribute named '%{name}'`, issue.HF{`activity`: issue.Label})
	issue.Hard2(WF_TOO_MANY_GUARDS, `%{activity} is too complex. Expected %{max} guards maximum. Have %{actual}`, issue.HF{`activity`: issue.Label})
	issue.Hard2(WF_PARAMETER_UNRESOLVED, `%{activity}, parameter %{parameter} cannot be resolved`, issue.HF{`activity`: issue.Label})
	issue.Hard(WF_OUTPUT_NOT_STRUCT, `expected activity to return a struct, got %{type}`)
	issue.Hard(WF_MISSING_REQUIRED_BLOCK, `action %{name} is missing required block %{block}`)
}
