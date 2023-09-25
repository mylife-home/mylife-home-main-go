package cmd

import (
	"encoding/json"
	"mylife-home-core/pkg/plugins"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"mylife-home-common/bus" // tmp
	"mylife-home-common/components"
	"mylife-home-common/config"
	"mylife-home-common/defines"
	"mylife-home-common/instance_info"
	"mylife-home-common/log"
	"mylife-home-common/tools"
)

var logger = log.CreateLogger("mylife:home:core:main")

var configFile string
var logConsole bool

var rootCmd = &cobra.Command{
	Use:   "mylife-home-core",
	Short: "mylife-home-core - Mylife Home Core",
	Run: func(_ *cobra.Command, _ []string) {
		defines.Init("core")
		log.Init(logConsole)
		config.Init(configFile)
		plugins.Build()
		instance_info.Init()

		logger.WithError(errors.Errorf("failed")).Error("bam")

		testComponent()
		transport := testBus()
		testRegistry(transport)
		testExit()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "config.yaml", "config file (default is $(PWD)/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&logConsole, "log-console", false, "Log to console")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func testComponent() {

	plugin := plugins.GetPlugin("logic-base.value-binary")

	logger.Infof("Metadata = '%s'", plugin.Metadata())

	comp, err := plugin.Instantiate("test", map[string]any{"initialValue": true})
	if err != nil {
		panic(err)
	}

	comp.SetOnStateChange(func(name string, value any) {
		logger.Infof("State '%s' changed to '%v'", name, value)
	})

	logger.Infof("State = '%v'", comp.GetState())

	logger.Infof("Execute")
	comp.Execute("setValue", false)

	logger.Infof("Execute no change")
	comp.Execute("setValue", false)

	logger.Infof("Terminate")
	comp.Termainte()
}

const pluginMeta = `
{
  "name": "value-binary",
  "module": "logic-base",
  "usage": "logic",
  "version": "42.42.42",
  "config": {},
  "members": {
    "value": { "memberType": "state", "valueType": "bool" },
    "setValue": { "memberType": "action", "valueType": "bool" }
  }
}
`

const compMeta = `
{"id":"test2","plugin":"logic-base.value-binary"}
`

func testBus() *bus.Transport {
	options := bus.NewOptions().SetPresenceTracking(true)
	transport := bus.NewTransport(options)

	transport.Presence().OnInstanceChange().Register(func(ipc *bus.InstancePresenceChange) {
		logger.Infof("%s online=%t", ipc.InstanceName(), ipc.Online())
	})

	waitOnline(transport)
	testLocalComp(transport)
	testRemoteComp(transport)
	testRpc(transport)

	return transport
}

func waitOnline(transport *bus.Transport) {
	onlinec := make(chan struct{}, 1)

	token := transport.OnOnlineChanged().Register(func(online bool) {
		if online {
			onlinec <- struct{}{}
		}
	})

	<-onlinec

	transport.OnOnlineChanged().Unregister(token)
}

func testLocalComp(transport *bus.Transport) {
	plugin := plugins.GetPlugin("logic-base.value-binary")

	comp, err := plugin.Instantiate("test2", map[string]any{"initialValue": false})
	if err != nil {
		panic(err)
	}

	transport.Metadata().Set("plugins/logic-base.value-binary", json.RawMessage(pluginMeta))
	transport.Metadata().Set("components/test2", json.RawMessage(compMeta))

	bcomp, err := transport.Components().AddLocalComponent("test2")
	if err != nil {
		panic(err)
	}

	bcomp.RegisterAction("setValue", func(value []byte) {
		comp.Execute("setValue", bus.Encoding.ReadBool(value))
	})

	comp.SetOnStateChange(func(name string, value any) {
		err := bcomp.SetState(name, bus.Encoding.WriteBool(value.(bool)))
		if err != nil {
			panic(err)
		}
	})

	state := comp.GetStateItem("value")
	err = bcomp.SetState("value", bus.Encoding.WriteBool(state.(bool)))
	if err != nil {
		panic(err)
	}

}

func testRemoteComp(transport *bus.Transport) {
	comp := transport.Components().TrackRemoteComponent(tools.Hostname()+"-core", "test2")
	comp.RegisterStateChange("value", func(value []byte) {
		logger.Infof("value set to %t", bus.Encoding.ReadBool(value))
	})

	val := false
	for i := 0; i < 4; i += 1 {
		time.Sleep(time.Second)

		logger.Infof("Set value to %t", val)
		comp.EmitAction("setValue", bus.Encoding.WriteBool(val))

		val = !val
	}
}

func testRpc(transport *bus.Transport) {

	type input struct {
		Message string `json:"message"`
	}

	type output string

	svc := bus.NewRpcService[input, output](func(i input) (output, error) {
		logger.Infof("serve input: %+v", i)
		return output(i.Message), nil
	})

	if err := transport.Rpc().Serve("test-service", svc); err != nil {
		panic(err)
	}

	out, err := bus.RpcCall[input, output](transport.Rpc(), tools.Hostname()+"-core", "test-service", input{Message: "toto"}, bus.RpcTimeout)
	if err != nil {
		panic(err)
	}

	logger.Infof("call result: %s", out)

	if err := transport.Rpc().Unserve("test-service"); err != nil {
		panic(err)
	}
}

func testRegistry(transport *bus.Transport) {
	options := components.NewRegistryOptions()
	options.PublishRemoteComponents(transport)
	reg := components.NewRegistry(options)

	time.Sleep(time.Second * 5)

	reg.Terminate()
}

func testExit() {

	exit := make(chan os.Signal, 1)

	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	s := <-exit
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	logger.Infof("Got signal %s", s)

}
