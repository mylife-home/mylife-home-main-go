package manager

import (
	"mylife-home-common/bus"
	"mylife-home-common/components"
	"mylife-home-common/instance_info"
	"mylife-home-common/log"
	"mylife-home-ui/pkg/model"
	"mylife-home-ui/pkg/web"
)

var logger = log.CreateLogger("mylife:home:ui:manager")

type Manager struct {
	transport *bus.Transport
	registry  components.Registry
	model     *model.ModelManager
	listener  components.BusListener
	api       *rpcApi
	webServer *web.WebServer
}

func MakeManager() *Manager {
	manager := &Manager{}

	manager.transport = bus.NewTransport()
	manager.registry = components.NewRegistry()
	manager.model = model.NewModelManager()
	manager.listener = components.ListenBus(manager.transport, manager.registry)
	manager.api = makeRpcApi(manager.transport, manager.model)
	manager.webServer = web.NewWebServer(manager.registry, manager.model)

	instance_info.AddCapability("ui-manager")

	return manager
}

func (manager *Manager) Terminate() {
	manager.webServer.Terminate()
	manager.api.Terminate()
	manager.listener.Terminate()
	manager.transport.Terminate()
}
