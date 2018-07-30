package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/puppetlabs/data-protobuf/misc"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/serialization"
	"github.com/puppetlabs/go-evaluator/types"
	// Ensure initialization of pcore
	_ "github.com/puppetlabs/go-evaluator/pcore"
	"github.com/puppetlabs/go-semver/semver"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Printf(`%s`, err)
		return
	}
	defer nc.Close()

	// Simple Publisher
	c := eval.Puppet.RootContext()
	v := types.WrapHash([]*types.HashEntry{
		types.WrapHashEntry2(`hi`, types.WrapString(`Hello`)),
		types.WrapHashEntry2(`what`, types.WrapHash([]*types.HashEntry{
			types.WrapHashEntry2(`where`, types.WrapString(`World`)),
			types.WrapHashEntry2(`version`, types.WrapSemVer(semver.MustParseVersion(`1.2.3`)))}))})

	tdc := serialization.NewToDataConverter(c, types.SingletonHash2(`rich_data`, types.Boolean_TRUE))
	pv, err := proto.Marshal(misc.ToPBData(tdc.Convert(v)))
	if err != nil {
		fmt.Printf(`%s`, err)
		return
	}
	nc.Publish("foo", pv)
}
