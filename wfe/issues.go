package wfe

import "github.com/lyraproj/issue/issue"

const (
	AlreadyDefined                 = `WF_ALREADY_DEFINED`
	ExpectedValueNotProduced       = `WF_EXPECTED_VALUE_NOT_PRODUCED`
	FailedToLoadPlugin             = `WF_FAILED_TO_LOAD_PLUGIN`
	GraphDotMarshal                = `WF_GRAPH_DOT_MARSHAL`
	IterationStepWrongParameters   = `WF_ITERATION_STEP_WRONG_PARAMETERS`
	IterationStepWrongReturns      = `WF_ITERATION_STEP_WRONG_OUTPUT`
	IterationParameterInvalidCount = `WF_ITERATION_PARAMETER_INVALID_COUNT`
	IterationParameterWrongType    = `WF_ITERATION_PARAMETER_WRONG_TYPE`
	IterationVariableInvalidCount  = `WF_ITERATION_VARIABLE_INVALID_COUNT`
	LyraLinkNoMap                  = `WF_LYRA_LINK_NO_MAP`
	LyraLinkNoExe                  = `WF_LYRA_LINK_NO_EXE`
	MultipleProducersOfValue       = `WF_MULTIPLE_PRODUCERS_OF_VALUE`
	NoProducerOfValue              = `WF_NO_PRODUCER_OF_VALUE`
	NoSuchAttribute                = `WF_NO_SUCH_ATTRIBUTE`
	NoStepContext                  = `WF_NO_STEP_CONTEXT`
	MissingRequiredProperty        = `WF_MISSING_REQUIRED_PROPERTY`
	MultipleErrors                 = `WF_MULTIPLE_ERRORS`
	ParameterUnresolved            = `WF_PARAMETER_UNRESOLVED`
	StepExecutionError             = `WF_STEP_EXECUTION_ERROR`
	TooManyGuards                  = `WF_TOO_MANY_GUARDS`
	UnableToLoadRequired           = `WF_UNABLE_TO_LOAD_REQUIRED`
	UnableToDetermineExternalId    = `WF_UNABLE_TO_DETERMINE_EXTERNAL_ID`
)

func init() {
	issue.Hard2(AlreadyDefined, `%{step} is already defined`, issue.HF{`step`: issue.Label})
	issue.Hard2(ExpectedValueNotProduced, `%{step} did not produce return value '%{value}'`, issue.HF{`step`: issue.Label})
	issue.Hard(FailedToLoadPlugin, `error while loading plugin executable '%{executable}': %{message}`)
	issue.Hard(GraphDotMarshal, `error while marshalling graph to dot: %{detail}`)
	issue.Hard2(IterationStepWrongParameters, `%{iterator} parameters must consume returns produced by the iterator`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(IterationStepWrongReturns, `%{iterator} returns must consist of a 'key' and a 'value'`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(IterationParameterInvalidCount, `%{iterator} wrong number of parameters. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(IterationParameterInvalidCount, `%{iterator} wrong number of parameters. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(IterationParameterWrongType, `%{iterator} parameter %{parameter} if of wrong type. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard2(IterationVariableInvalidCount, `%{iterator}, wrong number of variables. Expected %{expected}, actual %{actual}`, issue.HF{`iterator`: issue.Label})
	issue.Hard(LyraLinkNoMap, `Lyra Link did not contain a YAML map`)
	issue.Hard(LyraLinkNoExe, `Lyra Link did not contain a valid 'executable' entry`)
	issue.Hard2(MultipleProducersOfValue, `both %{step1} and %{step2} returns the value '%{value}'`, issue.HF{`step1`: issue.Label, `step2`: issue.Label})
	issue.Hard2(NoProducerOfValue, `%{step} value '%{value}' is never produced`, issue.HF{`step`: issue.Label})
	issue.Hard2(NoSuchAttribute, `%{step} has no attribute named '%{name}'`, issue.HF{`step`: issue.Label})
	issue.Hard(NoStepContext, `no step context was found in current scope`)
	issue.Hard(MissingRequiredProperty, `definition %{service} %{definition} is missing required property '%{key}'`)
	issue.Hard2(MultipleErrors, `multiple errors: %{errors}`, issue.HF{`errors`: issue.JoinErrors})
	issue.Hard2(ParameterUnresolved, `%{step}, parameter %{parameter} cannot be resolved`, issue.HF{`step`: issue.Label})
	issue.Hard2(StepExecutionError, `error while executing %{step}`, issue.HF{`step`: issue.Label})
	issue.Hard2(TooManyGuards, `%{step} is too complex. Expected %{max} guards maximum. Have %{actual}`, issue.HF{`step`: issue.Label})
	issue.Hard(UnableToLoadRequired, `unable to load required %{namespace} '%{name}'`)
	issue.Hard(UnableToDetermineExternalId, `unable to determine external ID for %{style} '%{id}'`)
}
