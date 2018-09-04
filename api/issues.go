package api

import "github.com/puppetlabs/go-issues/issue"

const (
	GENESIS_NO_SUCH_ACTION          = `GENESIS_NO_SUCH_ACTION`
	GENESIS_NOT_PUPPET_CONTEXT      = `GENESIS_NOT_PUPPET_CONTEXT`
	GENESIS_NOT_STRING_HASH         = `GENESIS_NOT_STRING_HASH`
	GENESIS_ACTION_NOT_STRUCTPTR    = `GENESIS_ACTION_NOT_STRUCTPTR`
	GENESIS_ACTION_NOT_FUNCTION     = `GENESIS_ACTION_NOT_FUNCTION`
	GENESIS_ACTION_NOT_STRUCT       = `GENESIS_ACTION_NOT_STRUCT`
	GENESIS_ACTION_BAD_FUNCTION     = `GENESIS_ACTION_BAD_FUNCTION`
	GENESIS_ACTION_BAD_RETURN_COUNT = `GENESIS_ACTION_BAD_RETURN_COUNT`
	GENESIS_ACTION_VALUE_NOT_REFLECTABLE = `GENESIS_ACTION_VALUE_NOT_REFLECTABLE`
)

func init() {
	issue.Hard(GENESIS_NO_SUCH_ACTION, `actor '%{actor}' has no action named '%{action}'`)
	issue.Hard(GENESIS_NOT_PUPPET_CONTEXT, `genesis context is not a puppet context`)
	issue.Hard(GENESIS_ACTION_NOT_FUNCTION, `expected Action '%{name}' to be a producerAction, got %{type}`)
	issue.Hard(GENESIS_ACTION_NOT_STRUCTPTR, `expected '%{name}' to be pointer to struct, got %{type}`)
	issue.Hard(GENESIS_ACTION_BAD_FUNCTION, `expected Action '%{name}' to be a producerAction of type func(fsm.Context, optional struct) (optional struct, error), got %{type}`)
	issue.Hard(GENESIS_NOT_STRING_HASH, `expected argument of type Hash[String,Any], got %{type}`)
	issue.Hard(GENESIS_ACTION_NOT_STRUCT, `expected Action '%{name}' to return a struct, got %{type}`)
	issue.Hard(GENESIS_ACTION_BAD_RETURN_COUNT, `expected Action '%{name}' to return %{expected_count} values, got %{actual_count}`)
	issue.Hard(GENESIS_ACTION_VALUE_NOT_REFLECTABLE, `value with key '%{key}', returned from Action '%{name}', has non-reflectable type %{type}`)
}

