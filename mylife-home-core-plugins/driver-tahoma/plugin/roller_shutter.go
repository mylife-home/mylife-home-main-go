package plugin

import (
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-tahoma/engine"
	"reflect"
)

// @Plugin(description="Volet roulant Somfy" usage="actuator")
type RollerShutter struct {

	// @Config(description="Identifiant de la box Somfy à partir de laquelle se connecter")
	BoxKey string

	// @Config(description="URL du périphérique Somfy")
	DeviceURL string

	// @State()
	Online definitions.State[bool]

	// @State()
	Exec definitions.State[bool]

	// @State(type="range[0;100]")
	Value definitions.State[int64]

	store                   *engine.Store
	storeOnlineChangedToken tools.RegistrationToken
	storeDeviceChangedToken tools.RegistrationToken
	storeStateChangedToken  tools.RegistrationToken
	storeExecChangedToken   tools.RegistrationToken
}

func (component *RollerShutter) Init() error {
	component.store = engine.GetStore(component.BoxKey)

	component.storeOnlineChangedToken = component.store.OnOnlineChanged().Register(component.handleOnlineChanged)
	component.storeDeviceChangedToken = component.store.OnDeviceChanged().Register(component.handleDeviceChanged)
	component.storeStateChangedToken = component.store.OnStateChanged().Register(component.handleStateChanged)
	component.storeExecChangedToken = component.store.OnExecChanged().Register(component.handleExecChanged)

	component.refreshOnline()

	return nil
}

func (component *RollerShutter) Terminate() {
	component.store.OnOnlineChanged().Unregister(component.storeOnlineChangedToken)
	component.store.OnDeviceChanged().Unregister(component.storeDeviceChangedToken)
	component.store.OnStateChanged().Unregister(component.storeStateChangedToken)
	component.store.OnExecChanged().Unregister(component.storeExecChangedToken)

	component.store = nil
	engine.ReleaseStore(component.BoxKey)
}

// @Action()
func (component *RollerShutter) DoOpen(arg bool) {
	if component.Online.Get() && arg {
		component.store.Execute(component.DeviceURL, "open")
	}
}

// @Action()
func (component *RollerShutter) DoClose(arg bool) {
	if component.Online.Get() && arg {
		component.store.Execute(component.DeviceURL, "close")
	}
}

// @Action()
func (component *RollerShutter) Toggle(arg bool) {
	if component.Online.Get() && arg {
		var cmd string
		if component.Value.Get() < 50 {
			cmd = "open"
		} else {
			cmd = "close"
		}

		component.store.Execute(component.DeviceURL, cmd)
	}
}

// @Action(type="range[-1;100]")
func (component *RollerShutter) SetValue(arg int64) {
	if component.Online.Get() && arg != -1 {
		component.store.Execute(component.DeviceURL, "setClosure", 100-arg)
	}
}

// @Action()
func (component *RollerShutter) Interrupt(arg bool) {
	if component.Online.Get() && arg {
		component.store.Interrupt(component.DeviceURL)
	}
}

func (component *RollerShutter) handleOnlineChanged(online bool) {
	component.refreshOnline()
}

func (component *RollerShutter) handleDeviceChanged(arg *engine.DeviceChange) {
	component.refreshOnline()
}

func (component *RollerShutter) handleStateChanged(state *engine.DeviceState) {
	if state.DeviceURL() != component.DeviceURL || state.Name() != "core:ClosureState" {
		return
	}

	if state.Type() != engine.StateInt {
		logger.Warnf("Invalid state type %d , expected %d (device='%s', state name='%s')", state.Type(), engine.StateInt, state.DeviceURL(), state.Name())
		return
	}

	value, ok := state.Value().(int64)
	if !ok {
		logger.Warnf("Could not cast value %+v of type %s to int (device='%s', state name='%s')", state.Value(), reflect.TypeOf(state.Value()), state.DeviceURL(), state.Name())
		return
	}

	if value < 0 {
		logger.Warnf("Invalid value %d, will use 0 (device='%s', state name='%s')", value, state.DeviceURL(), state.Name())
		value = 0
	}

	if value > 100 {
		logger.Warnf("Invalid value %d, will use 100 (device='%s', state name='%s')", value, state.DeviceURL(), state.Name())
		value = 100
	}

	// Note: somfy use reverse state compared to us
	component.Value.Set(100 - value)
}

func (component *SlidingGate) handleExecChanged(arg *engine.ExecChange) {
	if arg.DeviceURL() == component.DeviceURL {
		component.Exec.Set(arg.Executing())
	}
}

func (component *RollerShutter) refreshOnline() {
	online := component.store.Online() && component.store.GetDevice(component.DeviceURL) != nil

	component.Online.Set(online)

	if !online {
		component.Value.Set(0)
		component.Exec.Set(false)
	}
}
