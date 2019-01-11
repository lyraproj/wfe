package wfe

import "github.com/lyraproj/issue/issue"

const (
	WF_GRAPH_DOT_MARSHAL                 = `WF_GRAPH_DOT_MARSHAL`
	WF_ALREADY_DEFINED                   = `WF_ALREADY_DEFINED`
	WF_MULTIPLE_PRODUCERS_OF_VALUE       = `WF_MULTIPLE_PRODUCERS_OF_VALUE`
	WF_NO_PRODUCER_OF_VALUE              = `WF_NO_PRODUCER_OF_VALUE`
	WF_ITERATION_ACTIVITY_WRONG_INPUT    = `WF_ITERATION_ACTIVITY_WRONG_INPUT`
	WF_ITERATION_ACTIVITY_WRONG_OUTPUT   = `WF_ITERATION_ACTIVITY_WRONG_OUTPUT`
	WF_ITERATION_PARAMETER_INVALID_COUNT = `WF_ITERATION_PARAMETER_INVALID_COUNT`
	WF_ITERATION_PARAMETER_WRONG_TYPE    = `WF_ITERATION_PARAMETER_WRONG_TYPE`
	WF_ITERATION_VARIABLE_INVALID_COUNT  = `WF_ITERATION_VARIABLE_INVALID_COUNT`
	WF_NO_SUCH_ATTRIBUTE                 = `WF_NO_SUCH_ATTRIBUTE`
	WF_PARAMETER_UNRESOLVED              = `WF_PARAMETER_UNRESOLVED`
	WF_TOO_MANY_GUARDS                   = `WF_TOO_MANY_GUARDS`
)

func init() {

	issue.Hard(WF_GRAPH_DOT_MARSHAL, `error while marshalling graph to dot: %{detail}`)
	issue.Hard2(WF_ALREADY_DEFINED, `%{activity} is already defined`, issue.HF{`activity`: issue.Label})
	issue.Hard2(WF_MULTIPLE_PRODUCERS_OF_VALUE, `both %{activity1} and %{activity2} output the value '%{value}'`, issue.HF{`activity1`: issue.Label, `activity2`: issue.Label})
	issue.Hard2(WF_NO_PRODUCER_OF_VALUE, `%{activity} value '%{value}' is never produced`, issue.HF{`activity`: issue.Label})

	issue.Hard2(WF_ITERATION_ACTIVITY_WRONG_INPUT, `%{iterator} input must consume output produced by the iterator`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_ACTIVITY_WRONG_OUTPUT, `%{iterator} output must consist of a 'key' and a 'value'`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_PARAMETER_INVALID_COUNT, `%{iterator} wrong number of parameters. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_PARAMETER_INVALID_COUNT, `%{iterator} wrong number of parameters. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_PARAMETER_WRONG_TYPE, `%{iterator} parameter %{parameter} if of wrong type. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_ITERATION_VARIABLE_INVALID_COUNT, `%{iterator}, wrong number of variables. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(WF_NO_SUCH_ATTRIBUTE, `%{activity} has no attribute named '%{name}'`, issue.HF{`activity`: issue.Label})
	issue.Hard2(WF_TOO_MANY_GUARDS, `%{activity} is too complex. Expected %{max} guards maximum. Have %{actual}`, issue.HF{`activity`: issue.Label})
	issue.Hard2(WF_PARAMETER_UNRESOLVED, `%{activity}, parameter %{parameter} cannot be resolved`, issue.HF{`activity`: issue.Label})
}
