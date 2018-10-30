package activity

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-issues/issue"
)

type Operation int

const Create Operation = 0
const Read Operation = 1
const Update Operation = 2
const Delete Operation = 3
const Upsert Operation = 4

func (is Operation) String() string {
	switch is {
	case Create:
		return `create`
	case Read:
		return `read`
	case Update:
		return `update`
	case Upsert:
		return `upcert`
	case Delete:
		return `delete`
	default:
		return `unknown iteration style`
	}
}

func NewOperation(operation string) Operation {
	switch operation {
	case `create`:
		return Create
	case `read`:
		return Read
	case `update`:
		return Update
	case `upsert`:
		return Update
	case `detete`:
		return Delete
	}
	panic(eval.Error(WF_ILLEGAL_OPERATION, issue.H{`operation`: operation}))
}
