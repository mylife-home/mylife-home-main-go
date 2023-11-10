package plugin

import (
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-tahoma/engine"
	"sync"
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

	store                  *engine.Store
	storeOnlineChangedChan chan bool
	storeDeviceChangedChan chan *engine.DeviceChange
	storeExecChangedChan   chan *engine.ExecChange

	storeOnline bool
	device      *engine.Device
	onlineMux   sync.Mutex
}

func (component *SlidingGate) Init(runtime definitions.Runtime) error {
	component.store = engine.GetStore(component.BoxKey)

	component.storeOnlineChangedChan = make(chan bool)
	component.storeDeviceChangedChan = make(chan *engine.DeviceChange)
	component.storeExecChangedChan = make(chan *engine.ExecChange)

	tools.DispatchChannel(component.storeOnlineChangedChan, component.handleOnlineChanged)
	tools.DispatchChannel(component.storeDeviceChangedChan, component.handleDeviceChanged)
	tools.DispatchChannel(component.storeExecChangedChan, component.handleExecChanged)

	component.store.OnOnlineChanged().Subscribe(component.storeOnlineChangedChan, true)
	component.store.OnDeviceChanged().Subscribe(component.storeDeviceChangedChan)
	component.store.OnExecChanged().Subscribe(component.storeExecChangedChan)

	return nil
}

func (component *SlidingGate) Terminate() {
	component.store.OnOnlineChanged().Unsubscribe(component.storeOnlineChangedChan)
	component.store.OnDeviceChanged().Unsubscribe(component.storeDeviceChangedChan)
	component.store.OnExecChanged().Unsubscribe(component.storeExecChangedChan)

	close(component.storeOnlineChangedChan)
	close(component.storeDeviceChangedChan)
	close(component.storeExecChangedChan)

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
	component.onlineMux.Lock()
	defer component.onlineMux.Unlock()

	component.storeOnline = online

	component.refreshOnline()
}

func (component *SlidingGate) handleDeviceChanged(arg *engine.DeviceChange) {
	device := arg.Device()
	if device.DeviceURL() != component.DeviceURL {
		return
	}

	component.onlineMux.Lock()
	defer component.onlineMux.Unlock()

	switch arg.Operation() {
	case engine.OperationAdd:
		component.device = device
	case engine.OperationRemove:
		component.device = nil
	}

	component.refreshOnline()
}

func (component *SlidingGate) refreshOnline() {
	online := component.storeOnline && component.device != nil

	component.Online.Set(online)

	if !online {
		component.Exec.Set(false)
	}
}

func (component *SlidingGate) handleExecChanged(arg *engine.ExecChange) {
	if arg.DeviceURL() == component.DeviceURL {
		component.Exec.Set(arg.Executing())
	}
}
