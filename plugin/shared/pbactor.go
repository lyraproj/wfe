package shared

import (
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"github.com/puppetlabs/go-fsm/api"
)

type PbActor interface {
	GetActions() []*fsmpb.Action

	InvokeAction(id int64, parameters *datapb.DataHash, genesis api.Genesis) *datapb.DataHash
}
