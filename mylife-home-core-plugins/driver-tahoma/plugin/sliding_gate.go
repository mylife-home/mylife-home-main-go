package plugin

import (
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-tahoma/engine"
)

// @Plugin(description="Portail coulissant Somfy" usage="actuator")
type SlidingGate struct {
	// @Config(description="Identifiant de la box Somfy à partir de laquelle se connecter")
	BoxKey string

	// @Config(name="deviceURL" description="URL du périphérique Somfy")
	DeviceURL string

	// @State()
	Online definitions.State[bool]

	// @State()
	Exec definitions.State[bool]

	store                   *engine.Store
	storeOnlineChangedToken tools.RegistrationToken
	storeDeviceChangedToken tools.RegistrationToken
	storeExecChangedToken   tools.RegistrationToken
}

func (component *SlidingGate) Init() error {
	component.store = engine.GetStore(component.BoxKey)

	component.storeOnlineChangedToken = component.store.OnOnlineChanged().Register(component.handleOnlineChanged)
	component.storeDeviceChangedToken = component.store.OnDeviceChanged().Register(component.handleDeviceChanged)
	component.storeExecChangedToken = component.store.OnExecChanged().Register(component.handleExecChanged)

	component.refreshOnline()

	return nil
}

func (component *SlidingGate) Terminate() {
	component.store.OnOnlineChanged().Unregister(component.storeOnlineChangedToken)
	component.store.OnDeviceChanged().Unregister(component.storeDeviceChangedToken)
	component.store.OnExecChanged().Unregister(component.storeExecChangedToken)

	component.store = nil
	engine.ReleaseStore(component.BoxKey)
}

// @Action
func (component *SlidingGate) DoOpen(arg bool) {
	if component.Online.Get() && arg {
		component.store.Execute(component.DeviceURL, "open")
	}
}

// @Action
func (component *SlidingGate) DoClose(arg bool) {
	if component.Online.Get() && arg {
		component.store.Execute(component.DeviceURL, "close")
	}
}

// @Action
func (component *SlidingGate) Interrupt(arg bool) {
	if component.Online.Get() && arg {
		component.store.Interrupt(component.DeviceURL)
	}
}

func (component *SlidingGate) handleOnlineChanged(online bool) {
	go component.refreshOnline()
}

func (component *SlidingGate) handleDeviceChanged(arg *engine.DeviceChange) {
	go component.refreshOnline()
}

func (component *SlidingGate) handleExecChanged(arg *engine.ExecChange) {
	if arg.DeviceURL() == component.DeviceURL {
		component.Exec.Set(arg.Executing())
	}
}

func (component *SlidingGate) refreshOnline() {
	component.Online.Set(component.store.Online() && component.store.GetDevice(component.DeviceURL) != nil)
}
