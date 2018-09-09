package yaml2ast

import (
	"strings"

	"github.com/puppetlabs/go-issues/issue"
)

const (
	WF_YAML_DUPLICATE_KEY              = `WF_YAML_DUPLICATE_KEY`
	WF_YAML_ILLEGAL_TYPE               = `WF_YAML_ELEMENT_MUST_BE_HASH`
	EVAL_ILLEGAL_VARIABLE_NAME         = `EVAL_ILLEGAL_VARIABLE_NAME`
	WF_YAML_RESOURCE_TYPE_MUST_BE_NAME = `WF_YAML_RESOURCE_TYPE_MUST_BE_NAME`
	WF_YAML_UNRECOGNIZED_TOP_CONSTRUCT = `WF_YAML_UNRECOGNIZED_TOP_CONSTRUCT`
)

func joinPath(path interface{}) string {
	return strings.Join(path.([]string), `/`)
}

func init() {
	issue.Hard(WF_YAML_DUPLICATE_KEY, `the key '%{key}' is defined more than once. Path %{path}`)

	issue.Hard2(WF_YAML_ILLEGAL_TYPE, `the value must be %{expected}. Got %{actual}. Path %{path}`,
		issue.HF{`path`: joinPath, `expected`: issue.A_an, `actual`: issue.A_an})

	issue.Hard(EVAL_ILLEGAL_VARIABLE_NAME, `'%{name}' is not a legal variable name`)

	issue.Hard2(WF_YAML_RESOURCE_TYPE_MUST_BE_NAME, `'%{key}' is not a valid resource name. Path %{path}`, issue.HF{`path`: joinPath})

	issue.Hard2(WF_YAML_UNRECOGNIZED_TOP_CONSTRUCT, `unrecognized key '%{key}'. Path %{path}`, issue.HF{`path`: joinPath})
}
