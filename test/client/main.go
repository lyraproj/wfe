package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-fsm/test/common"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"github.com/puppetlabs/go-fsm/plugin/client"
)

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"actor": &client.Actor{},
}

func main() {
	log.SetOutput(ioutil.Discard)

	// We're a host. Start by launching the plugin process.
	projectHome := os.Getenv("GOPATH") + `/src/github.com/puppetlabs/go-fsm`
	pClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: common.Handshake,
		Plugins:         PluginMap,
//		Cmd:             exec.Command("/home/thhal/tools/node-v10.7.0-linux-x64/bin/node", "/home/thhal/git/genesis-js/src/genesis.js"),
		Cmd:             exec.Command(projectHome + "/test/bin/build_test_server"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})
	defer pClient.Kill()
	client.RunActions(pClient)
}
