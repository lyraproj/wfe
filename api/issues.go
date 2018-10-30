package api

import "github.com/puppetlabs/go-issues/issue"

const (
	WF_ILLEGAL_ITERATION_STYLE     = `WF_ILLEGAL_ITERATION_STYLE`
	WF_ILLEGAL_OPERATION           = `WF_ILLEGAL_OPERATION`
	WF_NO_SUCH_ACTIVITY            = `WF_NO_SUCH_ACTIVITY`
	WF_UNSUPPORTED_ACTIVITY_STYLE  = `WF_UNSUPPORTED_ACTIVITY_STYLE`
	GENESIS_ACTION_INVALID_ITERATE = `GENESIS_ACTION_INVALID_ITERATE`
	GENESIS_ACTION_NOT_STRUCT      = `GENESIS_ACTION_NOT_STRUCT`
	GENESIS_MULTI_ACTION_NOT_TUPLE = `GENESIS_MULTI_ACTION_NOT_TUPLE`
	WF_NO_ACTIVITY_CONTEXT         = `WF_NO_ACTIVITY_CONTEXT`
)

func init() {
	issue.Hard(WF_ILLEGAL_ITERATION_STYLE, `no such iteration style '%{style}'`)
	issue.Hard(WF_ILLEGAL_OPERATION, `no such operation '%{operation}'`)

	issue.Hard2(WF_NO_SUCH_ACTIVITY, `'%{workflow}' has no '%{activity}'`, issue.HF{`workflow`: issue.Label, `activity`: issue.Label})
	issue.Hard(WF_UNSUPPORTED_ACTIVITY_STYLE, `the activity style '%{style}' is not supported by this runtime`)
	issue.Hard2(GENESIS_ACTION_INVALID_ITERATE, `%{expression} cannot be used as an iterate expression`, issue.HF{`expression`: issue.A_an})
	issue.Hard(GENESIS_ACTION_NOT_STRUCT, `expected Activity '%{name}' to return a struct, got %{type}`)
	issue.Hard(GENESIS_MULTI_ACTION_NOT_TUPLE, `expected multi Activity '%{name}' to return a 2-element Tuple, got %{type}`)
	issue.Hard(WF_NO_ACTIVITY_CONTEXT, `no activity context was found in current scope`)
}
