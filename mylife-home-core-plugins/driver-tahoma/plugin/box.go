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

	store                   *engine.Store
	client                  *engine.Client
	storeOnlineChangedToken tools.RegistrationToken
}

func (component *Box) Init() error {
	component.store = engine.GetStore(component.BoxKey)
	component.storeOnlineChangedToken = component.store.OnOnlineChanged().Register(component.handleOnlineChanged)

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

	component.store.OnOnlineChanged().Unregister(component.storeOnlineChangedToken)
	engine.ReleaseStore(component.BoxKey)
}

func (component *Box) handleOnlineChanged(value bool) {
	component.Online.Set(value)
}
