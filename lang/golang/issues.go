package golang

import "github.com/puppetlabs/go-issues/issue"

const (
	WF_NOT_STRUCTPTR    = `WF_NOT_STRUCTPTR`
)

func init() {
	issue.Hard(WF_NOT_STRUCTPTR, `expected '%{name}' to be pointer to struct, got %{type}`)
}
