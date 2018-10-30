package typegen

import (
	"bytes"
	"github.com/puppetlabs/go-evaluator/eval"
)

type Generator interface {
	GenerateTypes(ts eval.TypeSet, ns []string, indent int, bld *bytes.Buffer)

	GenerateType(t eval.Type, ns []string, indent int, bld *bytes.Buffer)
}
