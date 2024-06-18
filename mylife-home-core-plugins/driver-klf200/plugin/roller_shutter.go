package plugin

import (
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-klf200/engine"
	"sync"
)

// @Plugin(description="Volet roulant Somfy" usage="actuator")
type RollerShutter struct {

	// @Config(description="Identifiant de la box KLF200 à partir de laquelle se connecter")
	BoxKey string

	// @Config(description="Nom saisi pour le périphérique sur la box KLF200")
	DeviceName string

	// @State()
	Online definitions.State[bool]

	// @State()
	Exec definitions.State[bool]

	// @State(type="range[0;100]")
	Value definitions.State[int64]

	store                  *engine.Store
	storeOnlineChangedChan chan bool
	storeDeviceChangedChan chan *engine.DeviceChange
	storeStateChangedChan  chan *engine.DeviceState

	storeOnline bool
	device      *engine.Device
	onlineMux   sync.Mutex
}

// TODO: handle Exec
// TODO: add step up/down

func (component *RollerShutter) Init(runtime definitions.Runtime) error {
	component.store = engine.GetStore(component.BoxKey)

	component.storeOnlineChangedChan = make(chan bool)
	component.storeDeviceChangedChan = make(chan *engine.DeviceChange)
	component.storeStateChangedChan = make(chan *engine.DeviceState)

	tools.DispatchChannel(component.storeOnlineChangedChan, component.handleOnlineChanged)
	tools.DispatchChannel(component.storeDeviceChangedChan, component.handleDeviceChanged)
	tools.DispatchChannel(component.storeStateChangedChan, component.handleStateChanged)

	component.store.Online().Subscribe(component.storeOnlineChangedChan, true)
	component.store.OnDeviceChanged().Subscribe(component.storeDeviceChangedChan)
	component.store.OnStateChanged().Subscribe(component.storeStateChangedChan)

	device := component.store.GetDevice(component.DeviceName)
	if device != nil {
		component.handleDeviceChanged(&engine.DeviceChange{})
	}

	return nil
}

func (component *RollerShutter) Terminate() {

	component.store.Online().Unsubscribe(component.storeOnlineChangedChan)
	component.store.OnDeviceChanged().Unsubscribe(component.storeDeviceChangedChan)
	component.store.OnStateChanged().Unsubscribe(component.storeStateChangedChan)

	close(component.storeOnlineChangedChan)
	close(component.storeDeviceChangedChan)
	close(component.storeStateChangedChan)

	component.store = nil
	engine.ReleaseStore(component.BoxKey)
}

// @Action()
func (component *RollerShutter) DoOpen(arg bool) {
	dev := component.device
	if component.Online.Get() && arg && dev != nil {
		component.store.Execute(dev.Index(), engine.MakeChangeAbsoluteCommand(100))
	}
}

// @Action()
func (component *RollerShutter) DoClose(arg bool) {
	dev := component.device
	if component.Online.Get() && arg && dev != nil {
		component.store.Execute(dev.Index(), engine.MakeChangeAbsoluteCommand(0))
	}
}

// @Action()
func (component *RollerShutter) Toggle(arg bool) {
	dev := component.device
	if component.Online.Get() && arg && dev != nil {
		component.store.Execute(dev.Index(), engine.MakeModeCommand())
	}
}

// @Action(type="range[-1;100]")
func (component *RollerShutter) SetValue(arg int64) {
	dev := component.device
	if component.Online.Get() && arg != -1 && dev != nil {
		component.store.Execute(dev.Index(), engine.MakeChangeAbsoluteCommand(int(arg)))
	}
}

// @Action()
func (component *RollerShutter) Interrupt(arg bool) {
	dev := component.device
	if component.Online.Get() && arg && dev != nil {
		component.store.Execute(dev.Index(), engine.MakeStopCommand())
	}
}

func (component *RollerShutter) handleOnlineChanged(online bool) {
	component.onlineMux.Lock()
	defer component.onlineMux.Unlock()

	component.storeOnline = online

	component.refreshOnline()
}

func (component *RollerShutter) handleDeviceChanged(arg *engine.DeviceChange) {
	device := arg.Device()
	if device.Name() != component.DeviceName {
		return
	}

	component.onlineMux.Lock()
	defer component.onlineMux.Unlock()

	switch arg.Operation() {
	case engine.OperationAdd:

		component.device = device

		go func() {
			// avoid deadlock
			state := component.store.GetState(device.Index())
			if state != nil {
				component.handleStateChanged(state)
			}
		}()

	case engine.OperationRemove:

		component.device = nil
	}

	component.refreshOnline()
}

func (component *RollerShutter) refreshOnline() {
	online := component.storeOnline && component.device != nil

	component.Online.Set(online)

	if !online {
		component.Value.Set(0)
		component.Exec.Set(false)
	}
}

func (component *RollerShutter) handleStateChanged(state *engine.DeviceState) {
	device := component.device

	if device == nil || state.DeviceIndex() != device.Index() {
		return
	}

	component.Value.Set(int64(state.CurrentPosition()))
}
