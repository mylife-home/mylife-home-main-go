package manager

import (
	"mylife-home-common/bus"
	"mylife-home-common/components"
	"mylife-home-common/log"
	"mylife-home-ui/pkg/model"
	"mylife-home-ui/pkg/web"
)

var logger = log.CreateLogger("mylife:home:ui:manager")

type Manager struct {
	transport *bus.Transport
	registry  components.Registry
	model     *model.ModelManager
	webServer *web.WebServer
}

func MakeManager() *Manager {
	manager := &Manager{}

	// static?
	manager.transport = bus.NewTransport()
	manager.registry = components.NewRegistry()
	manager.model = model.NewModelManager()
	manager.webServer = web.NewWebServer(manager.registry, manager.model)

	return manager
}

func (manager *Manager) Terminate() {
	manager.webServer.Terminate()
	manager.transport.Terminate()
}
