package plugin

import (
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-tahoma/engine"
)

// @Plugin(usage="actuator")
type Box struct {
	// @Config(description="Identifiant pour que les composants soient mises à jour à partir de cette box Somfy")
	BoxKey string

	// @Config()
	User string

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
	component.store.OnOnlineChanged().Subscribe(component.storeOnlineChangedChan, true)

	client, err := engine.MakeClient(component.User, component.Password)
	if err != nil {
		logger.WithError(err).Error("Error at client init")
		return nil
	}

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

	component.store.OnOnlineChanged().Unsubscribe(component.storeOnlineChangedChan)
	close(component.storeOnlineChangedChan)

	engine.ReleaseStore(component.BoxKey)
}

func (component *Box) handleOnlineChanged(value bool) {
	component.Online.Set(value)
}
