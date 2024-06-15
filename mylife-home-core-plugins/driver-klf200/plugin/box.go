package plugin

import (
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-klf200/engine"
)

// @Plugin(usage="actuator")
type Box struct {
	// @Config(description="Identifiant pour que les composants soient mises à jour à partir de cette box KLF200")
	BoxKey string

	// @Config(description="Format : 'IP/hostname:port'")
	Address string

	// @Config()
	Password string

	// @State()
	Online definitions.State[bool]

	store                  *engine.Store
	client                 *engine.Client
	storeOnlineChangedChan chan bool
}

func (component *Box) Init(runtime definitions.Runtime) error {
	component.store = engine.GetStore(component.BoxKey)

	component.storeOnlineChangedChan = make(chan bool)
	tools.DispatchChannel(component.storeOnlineChangedChan, component.handleOnlineChanged)
	component.store.Online().Subscribe(component.storeOnlineChangedChan, true)

	client := engine.MakeClient(component.Address, component.Password)

	component.client = client
	component.store.SetClient(component.client)

	return nil
}

func (component *Box) Terminate() {

	if component.client != nil {
		component.store.UnsetClient()
		component.client.Terminate()
		component.client = nil
	}

	component.store.Online().Unsubscribe(component.storeOnlineChangedChan)
	close(component.storeOnlineChangedChan)

	engine.ReleaseStore(component.BoxKey)
}

func (component *Box) handleOnlineChanged(value bool) {
	component.Online.Set(value)
}
