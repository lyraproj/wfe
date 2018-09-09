package golang

import "github.com/puppetlabs/go-issues/issue"

const (
	WF_BAD_FUNCTION     = `WF_BAD_FUNCTION`
	WF_BAD_RETURN_COUNT = `WF_BAD_RETURN_COUNT`
	WF_NOT_FUNCTION     = `WF_NOT_FUNCTION`
	WF_NOT_STRUCT       = `WF_NOT_STRUCT`
	WF_NOT_STRUCTPTR    = `WF_NOT_STRUCTPTR`
)

func init() {
	issue.Hard(WF_BAD_FUNCTION, `expected '%{name}' to be a producer function of type func(fsm.Context, optional struct) (optional struct, error), got %{type}`)
	issue.Hard(WF_BAD_RETURN_COUNT, `expected %{activity} to return %{expected_count} values, got %{actual_count}`)
	issue.Hard(WF_NOT_FUNCTION, `expected '%{name}' to be a function, got %{type}`)
	issue.Hard2(WF_NOT_STRUCT, `expected %{activity} to return a struct, got %{type}`, issue.HF{`activity`: issue.Label})
	issue.Hard(WF_NOT_STRUCTPTR, `expected '%{name}' to be pointer to struct, got %{type}`)
}
