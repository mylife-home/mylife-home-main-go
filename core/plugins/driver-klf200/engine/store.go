package engine

import (
	"mylife-home-common/tools"
	"sync"

	"github.com/gookit/goutil/errorx/panics"
	"github.com/mylife-home/klf200-go"
	"github.com/mylife-home/klf200-go/commands"
	"golang.org/x/exp/maps"
)

type Operation int

const (
	OperationAdd Operation = iota + 1
	OperationRemove
)

type DeviceChange struct {
	device    *Device
	operation Operation
}

func MakeDeviceChange(dev *Device, op Operation) *DeviceChange {
	return &DeviceChange{
		device:    dev,
		operation: op,
	}
}

func (change *DeviceChange) Device() *Device {
	return change.device
}

func (change *DeviceChange) Operation() Operation {
	return change.operation
}

type Store struct {
	client                  *Client
	clientOnlineChangedChan chan bool
	clientDevicesChan       chan []*Device
	clientStatesChan        chan []*klf200.StatusData
	clientMux               sync.Mutex

	devices         map[string]*Device   // key = name
	states          map[int]*DeviceState // key = node index
	online          tools.SubjectValue[bool]
	onDeviceChanged tools.Subject[*DeviceChange]
	onStateChanged  tools.Subject[*DeviceState]

	mux sync.Mutex
}

func newStore() *Store {
	return &Store{
		online:          tools.MakeSubjectValue(false),
		devices:         make(map[string]*Device),
		states:          make(map[int]*DeviceState),
		onDeviceChanged: tools.MakeSubject[*DeviceChange](),
		onStateChanged:  tools.MakeSubject[*DeviceState](),
	}
}

func (store *Store) SetClient(client *Client) {
	store.clientMux.Lock()
	defer store.clientMux.Unlock()

	panics.IsTrue(store.client == nil)

	store.client = client

	store.clientOnlineChangedChan = make(chan bool)
	store.clientDevicesChan = make(chan []*Device)
	store.clientStatesChan = make(chan []*klf200.StatusData)

	tools.DispatchChannel(store.clientOnlineChangedChan, store.handleOnlineChanged)
	tools.DispatchChannel(store.clientDevicesChan, store.handleDevicesChanged)
	tools.DispatchChannel(store.clientStatesChan, store.handleStatesChanged)

	store.client.Online().Subscribe(store.clientOnlineChangedChan, true)
	store.client.Devices().Subscribe(store.clientDevicesChan)
	store.client.States().Subscribe(store.clientStatesChan)
}

func (store *Store) clearClient() {
	store.clientMux.Lock()
	defer store.clientMux.Unlock()

	panics.IsTrue(store.client != nil)

	store.client.Online().Unsubscribe(store.clientOnlineChangedChan)
	store.client.Devices().Unsubscribe(store.clientDevicesChan)
	store.client.States().Unsubscribe(store.clientStatesChan)

	close(store.clientOnlineChangedChan)
	close(store.clientDevicesChan)
	close(store.clientStatesChan)

	store.clientOnlineChangedChan = nil
	store.clientDevicesChan = nil
	store.clientStatesChan = nil

	store.online.Update(false)

	store.client = nil
}

func (store *Store) UnsetClient() {
	store.clearClient()

	store.mux.Lock()
	defer store.mux.Unlock()
	store.clearDevices()
}

func (store *Store) Online() tools.ObservableValue[bool] {
	return store.online
}

func (store *Store) OnDeviceChanged() tools.Observable[*DeviceChange] {
	return store.onDeviceChanged
}

func (store *Store) OnStateChanged() tools.Observable[*DeviceState] {
	return store.onStateChanged
}

func (store *Store) GetDevice(deviceName string) *Device {
	store.mux.Lock()
	defer store.mux.Unlock()

	return store.devices[deviceName]
}

func (store *Store) GetState(deviceIndex int) *DeviceState {
	store.mux.Lock()
	defer store.mux.Unlock()

	return store.states[deviceIndex]
}

func (store *Store) handleOnlineChanged(online bool) {
	store.online.Update(online)
	// for now we consider devices stay even if offline (and states will stay accurate ...)
}

func (store *Store) handleDevicesChanged(devices []*Device) {
	store.mux.Lock()
	defer store.mux.Unlock()

	list := make(map[string]struct{})

	// add/update
	for _, device := range devices {
		name := device.Name()
		list[name] = struct{}{}
		existing := store.devices[name]

		if existing == nil {
			store.addDevice(device)
		}
	}

	// remove
	for _, device := range maps.Values(store.devices) {
		name := device.Name()
		if _, exists := list[name]; !exists {
			store.removeDevice(device)
		}
	}

}

func (store *Store) handleStatesChanged(states []*klf200.StatusData) {
	store.mux.Lock()
	defer store.mux.Unlock()

	for _, state := range states {
		store.refreshState(state)
	}
}

func (store *Store) addDevice(device *Device) {
	name := device.Name()

	store.devices[name] = device

	logger.Debugf("Device joined '%s' (%d)", name, device.Index())

	store.onDeviceChanged.Notify(MakeDeviceChange(device, OperationAdd))
}

func (store *Store) removeDevice(device *Device) {
	name := device.Name()

	delete(store.devices, name)
	delete(store.states, device.Index())

	logger.Debugf("Device left '%s' (%d)", name, device.Index())

	store.onDeviceChanged.Notify(MakeDeviceChange(device, OperationRemove))
}

func (store *Store) clearDevices() {
	for _, device := range maps.Values(store.devices) {
		delete(store.devices, device.Name())
		delete(store.states, device.Index())

		store.onDeviceChanged.Notify(MakeDeviceChange(device, OperationRemove))
	}
}

func (store *Store) refreshState(state *klf200.StatusData) {
	// compute position
	ok, currentPosition := state.CurrentPosition.Absolute()
	if !ok {
		logger.Warnf("Unsupported state value %v", state.CurrentPosition)
		return
	}

	// reverse order
	currentPosition = 100 - currentPosition

	// check if need to update
	deviceIndex := state.NodeIndex

	oldState := store.states[deviceIndex]

	if oldState == nil || oldState.CurrentPosition() != currentPosition {
		newState := &DeviceState{
			deviceIndex:     state.NodeIndex,
			currentPosition: currentPosition,
		}

		store.states[deviceIndex] = newState

		logger.Debugf("Device state changed: device %d => %d", newState.DeviceIndex(), newState.CurrentPosition())
		store.onStateChanged.Notify(newState)
	}
}

type Command interface {
	execute(client *Client, deviceIndex int) (*klf200.Session, error)
}

type modeCommand struct {
}

func (*modeCommand) execute(client *Client, deviceIndex int) (*klf200.Session, error) {
	return client.Mode(deviceIndex)
}

func MakeModeCommand() Command {
	return &modeCommand{}
}

type changePositionCommand struct {
	position commands.MPValue
}

func (cmd *changePositionCommand) execute(client *Client, deviceIndex int) (*klf200.Session, error) {
	return client.ChangePosition(deviceIndex, cmd.position)
}

func MakeChangeAbsoluteCommand(value int) Command {
	return &changePositionCommand{position: commands.NewMPValueAbsolute(100 - value)}
}

func MakeChangeRelativeCommand(value int) Command {
	return &changePositionCommand{position: commands.NewMPValueRelative(-value)}
}

func MakeStopCommand() Command {
	return &changePositionCommand{position: commands.NewMPValueCurrent()}
}

func (store *Store) Execute(deviceIndex int, command Command) {
	store.mux.Lock()
	defer store.mux.Unlock()

	client := store.client

	if client == nil {
		logger.Warn("Execute without client, ignored")
		return
	}

	session, err := command.execute(client, deviceIndex)

	if err != nil {
		logger.WithError(err).Errorf("could not execute command on device %d", deviceIndex)
		return
	}

	// TODO: use session to get execution data
	_ = session
}
