package manager

import (
	"mylife-home-common/bus"
	"mylife-home-common/components"
	"mylife-home-common/config"
	"mylife-home-common/instance_info"
	"mylife-home-common/log"
	"mylife-home-core/pkg/plugins"
)

var logger = log.CreateLogger("mylife:home:core:manager")

type managerConfig struct {
	SupportsBindings bool `mapstructure:"supportsBindings"`
}

type Manager struct {
	transport *bus.Transport
	registry  components.Registry
	cm        *componentManager
	api       *rpcApi
	publisher components.BusPublisher
	listener  components.BusListener
}

func MakeManager() *Manager {
	conf := managerConfig{}
	config.BindStructure("manager", &conf)

	supportsBindings := conf.SupportsBindings

	manager := &Manager{}

	// static?
	manager.addPluginsInstanceInfo()

	manager.transport = bus.NewTransport()
	manager.registry = components.NewRegistry()
	manager.cm = makeComponentManager(manager.registry, supportsBindings)
	manager.api = makeRpcApi(manager.transport, manager.cm, supportsBindings)
	manager.publisher = components.PublishBus(manager.transport, manager.registry)

	if supportsBindings {
		manager.listener = components.ListenBus(manager.transport, manager.registry)
	}

	return manager
}

func (manager *Manager) Terminate() {

	if manager.listener != nil {
		manager.listener.Terminate()
	}

	manager.publisher.Terminate()
	manager.api.Terminate()
	manager.cm.Terminate()
	//manager.registry.Terminate()
	manager.transport.Terminate()
}

func (manager *Manager) addPluginsInstanceInfo() {
	modules := make(map[string]string)
	for _, id := range plugins.Ids() {
		meta := plugins.GetPlugin(id).Metadata()
		modules[meta.Module()] = meta.Version()
	}

	for module, version := range modules {
		instance_info.AddComponent("core-plugins-"+module, version)
	}
}
