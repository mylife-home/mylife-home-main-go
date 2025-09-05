package manager

import (
	"mylife-home-common/bus"
	"mylife-home-common/components"
	"mylife-home-common/log"
	"mylife-home-ui/pkg/web"
)

var logger = log.CreateLogger("mylife:home:ui:manager")

type Manager struct {
	transport *bus.Transport
	registry  components.Registry
	webServer *web.WebServer
}

func MakeManager() *Manager {
	manager := &Manager{}

	// static?

	manager.transport = bus.NewTransport()
	manager.registry = components.NewRegistry()
	manager.webServer = web.NewWebServer(manager.registry)

	return manager
}

func (manager *Manager) Terminate() {
	manager.webServer.Terminate()
	//manager.registry.Terminate()
	manager.transport.Terminate()
}
