package api

import "github.com/lyraproj/issue/issue"

const (
	WF_ILLEGAL_ITERATION_STYLE         = `WF_ILLEGAL_ITERATION_STYLE`
	WF_ILLEGAL_OPERATION               = `WF_ILLEGAL_OPERATION`
	WF_NO_SUCH_ACTIVITY                = `WF_NO_SUCH_ACTIVITY`
	WF_NO_SUCH_ATTRIBUTE               = `WF_NO_SUCH_ATTRIBUTE`
	WF_UNSUPPORTED_ACTIVITY_STYLE      = `WF_UNSUPPORTED_ACTIVITY_STYLE`
	WF_ACTION_INVALID_ITERATE          = `WF_ACTION_INVALID_ITERATE`
	WF_ACTION_NOT_STRUCT               = `WF_ACTION_NOT_STRUCT`
	WF_MULTI_ACTION_NOT_TUPLE          = `WF_MULTI_ACTION_NOT_TUPLE`
	WF_NO_ACTIVITY_CONTEXT             = `WF_NO_ACTIVITY_CONTEXT`
	WF_MISSING_REQUIRED_PROPERTY       = `WF_MISSING_REQUIRED_PROPERTY`
	WF_UNABLE_TO_LOAD_REQUIRED         = `WF_UNABLE_TO_LOAD_REQUIRED`
	WF_UNABLE_TO_DETERMINE_EXTERNAL_ID = `WF_UNABLE_TO_DETERMINE_EXTERNAL_ID`
)

func init() {
	issue.Hard(WF_ILLEGAL_ITERATION_STYLE, `no such iteration style '%{style}'`)
	issue.Hard(WF_ILLEGAL_OPERATION, `no such operation '%{operation}'`)

	issue.Hard2(WF_NO_SUCH_ACTIVITY, `'%{workflow}' has no '%{activity}'`, issue.HF{`workflow`: issue.Label, `activity`: issue.Label})
	issue.Hard(WF_UNSUPPORTED_ACTIVITY_STYLE, `the activity style '%{style}' is not supported by this runtime`)
	issue.Hard2(WF_NO_SUCH_ATTRIBUTE, `%{activity} has no attribute named '%{name}'`, issue.HF{`activity`: issue.Label})
	issue.Hard2(WF_ACTION_INVALID_ITERATE, `%{expression} cannot be used as an iterate expression`, issue.HF{`expression`: issue.A_an})
	issue.Hard(WF_ACTION_NOT_STRUCT, `expected Activity '%{name}' to return a struct, got %{type}`)
	issue.Hard(WF_MULTI_ACTION_NOT_TUPLE, `expected multi Activity '%{name}' to return a 2-element Tuple, got %{type}`)
	issue.Hard(WF_NO_ACTIVITY_CONTEXT, `no activity context was found in current scope`)
	issue.Hard(WF_MISSING_REQUIRED_PROPERTY, `definition %{service} %{definition} is missing required property '%{key}'`)
	issue.Hard(WF_UNABLE_TO_LOAD_REQUIRED, `unable to load required %{namespace} '%{name}'`)
	issue.Hard(WF_UNABLE_TO_DETERMINE_EXTERNAL_ID, `unable to determine external ID for %{style} '%{id}'`)
}
