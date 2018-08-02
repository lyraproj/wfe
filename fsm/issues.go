package fsm

import "github.com/puppetlabs/go-issues/issue"

const (
	GENESIS_ACTION_ALREADY_DEFINED      = `GENESIS_ACTION_ALREADY_DEFINED`
	GENESIS_MULTIPLE_PRODUCERS_OF_VALUE = `GENESIS_MULTIPLE_PRODUCERS_OF_VALUE`
	GENESIS_NO_PRODUCER_OF_VALUE        = `GENESIS_NO_PRODUCER_OF_VALUE`
	GENESIS_UNABLE_TO_REFLECT_TYPE      = `GENESIS_UNABLE_TO_REFLECT_TYPE`
	GENESIS_ACTION_BAD_CONSUMES_COUNT   = `GENESIS_ACTION_BAD_CONSUMES_COUNT`
)

func init() {
	issue.Hard(GENESIS_ACTION_ALREADY_DEFINED, `action '%{name}' is already defined`)
	issue.Hard(GENESIS_MULTIPLE_PRODUCERS_OF_VALUE, `both '%{name1}' and '%{name2}' produces the value '%{value}'`)
	issue.Hard(GENESIS_NO_PRODUCER_OF_VALUE, `no action produces value '%{value}' required by action '%{action}'`)
	issue.Hard(GENESIS_UNABLE_TO_REFLECT_TYPE, `unable to convert type %{type} into a reflect.Type`)
	issue.Hard(GENESIS_ACTION_BAD_CONSUMES_COUNT, `action '%{name}' expects %{expected} parameters, got %{actual}`)
}
