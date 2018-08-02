package shared

import "github.com/puppetlabs/data-protobuf/datapb"

type PbGenesis interface {
	Apply(resources *datapb.DataHash) *datapb.DataHash
}
