package typegen

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"bytes"
)

type Generator interface {
	GenerateTypes(ts eval.TypeSet, ns []string, indent int, bld *bytes.Buffer)

	GenerateType(t eval.PType, ns []string, indent int, bld *bytes.Buffer)
}

