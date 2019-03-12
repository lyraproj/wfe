package wfe

import "github.com/lyraproj/issue/issue"

const (
	GraphDotMarshal                = `WF_GRAPH_DOT_MARSHAL`
	AlreadyDefined                 = `WF_ALREADY_DEFINED`
	MultipleProducersOfValue       = `WF_MULTIPLE_PRODUCERS_OF_VALUE`
	NoProducerOfValue              = `WF_NO_PRODUCER_OF_VALUE`
	IterationActivityWrongInput    = `WF_ITERATION_ACTIVITY_WRONG_INPUT`
	IterationActivityWrongOutput   = `WF_ITERATION_ACTIVITY_WRONG_OUTPUT`
	IterationParameterInvalidCount = `WF_ITERATION_PARAMETER_INVALID_COUNT`
	IterationParameterWrongType    = `WF_ITERATION_PARAMETER_WRONG_TYPE`
	IterationVariableInvalidCount  = `WF_ITERATION_VARIABLE_INVALID_COUNT`
	ParameterUnresolved            = `WF_PARAMETER_UNRESOLVED`
	TooManyGuards                  = `WF_TOO_MANY_GUARDS`
)

func init() {
	issue.Hard(GraphDotMarshal, `error while marshalling graph to dot: %{detail}`)
	issue.Hard2(AlreadyDefined, `%{activity} is already defined`, issue.HF{`activity`: issue.Label})
	issue.Hard2(MultipleProducersOfValue, `both %{activity1} and %{activity2} output the value '%{value}'`, issue.HF{`activity1`: issue.Label, `activity2`: issue.Label})
	issue.Hard2(NoProducerOfValue, `%{activity} value '%{value}' is never produced`, issue.HF{`activity`: issue.Label})
	issue.Hard2(IterationActivityWrongInput, `%{iterator} input must consume output produced by the iterator`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(IterationActivityWrongOutput, `%{iterator} output must consist of a 'key' and a 'value'`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(IterationParameterInvalidCount, `%{iterator} wrong number of parameters. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(IterationParameterInvalidCount, `%{iterator} wrong number of parameters. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(IterationParameterWrongType, `%{iterator} parameter %{parameter} if of wrong type. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(IterationVariableInvalidCount, `%{iterator}, wrong number of variables. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(TooManyGuards, `%{activity} is too complex. Expected %{max} guards maximum. Have %{actual}`, issue.HF{`activity`: issue.Label})
	issue.Hard2(ParameterUnresolved, `%{activity}, parameter %{parameter} cannot be resolved`, issue.HF{`activity`: issue.Label})
}
