package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-fsm/test/common"
	"io/ioutil"
	"log"
	"os/exec"
	"github.com/puppetlabs/go-fsm/plugin/client"
	"os"
)

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"actor": &client.Actor{},
}

func main() {
	log.SetOutput(ioutil.Discard)

	// We're a host. Start by launching the plugin process.
	// projectHome := os.Getenv("GOPATH") + `/src/github.com/puppetlabs/go-fsm`
	home := os.Getenv("HOME")
	pClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: common.Handshake,
		Plugins:         PluginMap,
		Cmd:             exec.Command(home + "/tools/node/bin/node", home + "/git/js-fsm/src/examples/ec2_attachinternetgw.js"),//
		// Cmd:             exec.Command(projectHome + "/test/bin/build_test_server"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})
	defer pClient.Kill()
	client.RunActions(pClient)
}
