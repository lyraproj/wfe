package api

import "github.com/puppetlabs/go-issues/issue"

const (
	GENESIS_ACTION_NOT_FUNCTION     = `GENESIS_ACTION_NOT_FUNCTION`
	GENESIS_ACTION_BAD_FUNCTION     = `GENESIS_ACTION_BAD_FUNCTION`
	GENESIS_ACTION_BAD_RETURN       = `GENESIS_ACTION_BAD_RETURN`
	GENESIS_ACTION_BAD_RETURN_COUNT = `GENESIS_ACTION_BAD_RETURN_COUNT`
)

func init() {
	issue.Hard(GENESIS_ACTION_NOT_FUNCTION, `expected action '%{name}' to be a function, got %{type}`)
	issue.Hard(GENESIS_ACTION_BAD_FUNCTION, `expected action '%{name}' to be a function of type func(fsm.Context, optional struct) (optional struct, error), got %{type}`)
	issue.Hard(GENESIS_ACTION_BAD_RETURN, `expected action '%{name}' to return a struct, got %{type}`)
	issue.Hard(GENESIS_ACTION_BAD_RETURN_COUNT, `expected action '%{name}' to return %{expected_count} values, got %{actual_count}`)
}
