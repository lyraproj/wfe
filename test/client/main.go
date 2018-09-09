package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-fsm/test/common"
	"io/ioutil"
	"log"
	"os/exec"
	"github.com/puppetlabs/go-fsm/lang/rpc/client"
	"os"
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"

	// Ensure Pcore is initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"actors": &client.ActorsPlugin{},
}

func main() {
	log.SetOutput(ioutil.Discard)

	// We're a host. Start by launching the plugin process.
	err := eval.Puppet.Do(func(ctx eval.Context) error {
		projectHome := os.Getenv("GOPATH") + `/src/github.com/puppetlabs/go-fsm`
		home := os.Getenv("HOME")
		jsClient := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: common.Handshake,
			Plugins:         PluginMap,
			Cmd:             exec.Command("/usr/local/bin/node", home + "/git/js-fsm/dist/examples/ec2_attachinternetgw.js"),
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		})
		defer jsClient.Kill()
		fmt.Println(client.RunActor(ctx, `attach`, jsClient, eval.EMPTY_MAP))

		goClient := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: common.Handshake,
			Plugins:         PluginMap,
			Cmd:             exec.Command(projectHome + "/test/bin/go_build_server"),
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		})
		defer goClient.Kill()

		fmt.Println(client.RunActor(ctx, `testing`, goClient, eval.EMPTY_MAP))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}
