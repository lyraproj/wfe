package shared

import (
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/fsmpb"
)

type PbActor interface {
	GetActions() []*fsmpb.Action

	InvokeAction(id int64, parameters *datapb.DataHash, genesis *PbGenesis) *datapb.DataHash
}
