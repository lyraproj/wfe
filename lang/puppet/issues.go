package puppet

import "github.com/puppetlabs/go-issues/issue"

const (
	WF_FIELD_TYPE_MISMATCH = `WF_FIELD_TYPE_MISMATCH`
	WF_ELEMENT_NOT_PARAMETER = `WF_ELEMENT_NOT_PARAMETER`
	WF_NO_DEFINITION = `WF_NO_DEFINITION`
	WF_NOT_ACTIVITY = `WF_NOT_ACTIVITY`
	WF_UNSUPPORTED_EXPRESSION = `WF_UNSUPPORTED_EXPRESSION`
)

func init() {
	issue.Hard(WF_FIELD_TYPE_MISMATCH, `expected activity %{field} to be a %{expected}, got %{actual}`)
	issue.Hard(WF_ELEMENT_NOT_PARAMETER, `expected activity %{field} element to be a Parameter, got %{type}`)
	issue.Hard(WF_NO_DEFINITION, `expected activity to contain a definition block`)
	issue.Hard2(WF_UNSUPPORTED_EXPRESSION, `%{expression} is not supported here`, issue.HF{`expression`: issue.A_an})
	issue.Hard2(WF_NOT_ACTIVITY, `block may only contain workflow activites. %{actual} is not supported here`,
		issue.HF{`actual`: issue.A_anUc})
}
