package api

import "github.com/lyraproj/issue/issue"

const (
	NoSuchAttribute             = `WF_NO_SUCH_ATTRIBUTE`
	NoActivityContext           = `WF_NO_ACTIVITY_CONTEXT`
	MissingRequiredProperty     = `WF_MISSING_REQUIRED_PROPERTY`
	MultipleErrors              = `WF_MULTIPLE_ERRORS`
	UnableToLoadRequired        = `WF_UNABLE_TO_LOAD_REQUIRED`
	UnableToDetermineExternalId = `WF_UNABLE_TO_DETERMINE_EXTERNAL_ID`
)

func init() {
	issue.Hard2(NoSuchAttribute, `%{activity} has no attribute named '%{name}'`, issue.HF{`activity`: issue.Label})
	issue.Hard(NoActivityContext, `no activity context was found in current scope`)
	issue.Hard(MissingRequiredProperty, `definition %{service} %{definition} is missing required property '%{key}'`)
	issue.Hard2(MultipleErrors, `multiple errors: %{errors}`, issue.HF{`errors`: issue.JoinErrors})
	issue.Hard(UnableToLoadRequired, `unable to load required %{namespace} '%{name}'`)
	issue.Hard(UnableToDetermineExternalId, `unable to determine external ID for %{style} '%{id}'`)
}
