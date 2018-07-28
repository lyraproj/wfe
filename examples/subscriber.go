package main

import (
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/nats-io/go-nats"
	"github.com/puppetlabs/go-evaluator/serialization"
	"github.com/golang/protobuf/proto"
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/data-protobuf/misc"
	"runtime"

	// Ensure initialization of pcore
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Printf(`%s`, err)
		return
	}

	// Simple Async Subscriber
	c := eval.Puppet.RootContext()
	fdc := serialization.NewFromDataConverter(c, eval.EMPTY_MAP)
	nc.Subscribe("foo", func(m *nats.Msg) {
		pv := &datapb.Data{}
		err = proto.Unmarshal(m.Data, pv)
		v := fdc.Convert(misc.FromPBData(pv))
		fmt.Printf("Received a message: %s\n", v)
	})
	runtime.Goexit()
}

