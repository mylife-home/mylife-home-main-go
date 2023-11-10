package engine

import (
	"mylife-home-common/tools"
	"reflect"
	"sync"

	"github.com/gookit/goutil/errorx/panics"
	"github.com/sgrimee/kizcool"
	"golang.org/x/exp/maps"
)

type Device struct {
	kiz kizcool.Device
}

func (dev *Device) DeviceURL() string {
	return string(dev.kiz.DeviceURL)
}

func (dev *Device) Type() string {
	return dev.kiz.Definition.QualifiedName
}

type DeviceStateType = kizcool.StateType

const (
	StateInt    DeviceStateType = kizcool.StateInt
	StateFloat  DeviceStateType = kizcool.StateFloat
	StateString DeviceStateType = kizcool.StateString
)

type DeviceState struct {
	deviceURL string
	name      string
	typ       DeviceStateType
	value     any
}

func (state *DeviceState) DeviceURL() string {
	return state.deviceURL
}

func (state *DeviceState) Name() string {
	return state.name
}

func (state *DeviceState) Type() DeviceStateType {
	return state.typ
}

func (state *DeviceState) Value() any {
	return state.value
}

type Operation int

const (
	OperationAdd Operation = iota + 1
	OperationRemove
)

type DeviceChange struct {
	device    *Device
	operation Operation
}

func (change *DeviceChange) Device() *Device {
	return change.device
}

func (change *DeviceChange) Operation() Operation {
	return change.operation
}

type ExecChange struct {
	deviceURL string
	executing bool
}

func (change *ExecChange) DeviceURL() string {
	return change.deviceURL
}

func (change *ExecChange) Executing() bool {
	return change.executing
}

type Store struct {
	client                  *Client
	clientOnlineChangedChan chan bool
	clientDeviceListChan    chan []kizcool.Device
	clientStateRefreshChan  chan *StateRefresh
	clientExecRefreshChan   chan *ExecChange
	clientMux               sync.Mutex

	devices         map[string]*Device
	states          map[string]*DeviceState // key = <deviceURL>$<name>
	onOnlineChanged tools.SubjectValue[bool]
	onDeviceChanged tools.Subject[*DeviceChange]
	onStateChanged  tools.Subject[*DeviceState]
	onExecChanged   tools.Subject[*ExecChange]

	mux sync.Mutex
}

func newStore() *Store {
	return &Store{
		devices:         make(map[string]*Device),
		states:          make(map[string]*DeviceState),
		onOnlineChanged: tools.MakeSubjectValue[bool](false),
		onDeviceChanged: tools.MakeSubject[*DeviceChange](),
		onStateChanged:  tools.MakeSubject[*DeviceState](),
		onExecChanged:   tools.MakeSubject[*ExecChange](),
	}
}

func (store *Store) SetClient(client *Client) {
	store.clientMux.Lock()
	defer store.clientMux.Unlock()

	panics.IsTrue(store.client == nil)

	store.client = client

	store.clientOnlineChangedChan = make(chan bool)
	store.clientDeviceListChan = make(chan []kizcool.Device)
	store.clientStateRefreshChan = make(chan *StateRefresh)
	store.clientExecRefreshChan = make(chan *ExecChange)

	tools.DispatchChannel(store.clientOnlineChangedChan, store.handleOnlineChanged)
	tools.DispatchChannel(store.clientDeviceListChan, store.handleDeviceList)
	tools.DispatchChannel(store.clientStateRefreshChan, store.handleStateRefresh)
	tools.DispatchChannel(store.clientExecRefreshChan, store.handleExecRefresh)

	store.client.OnOnlineChanged().Subscribe(store.clientOnlineChangedChan, true)
	store.client.OnDeviceList().Subscribe(store.clientDeviceListChan)
	store.client.OnStateRefresh().Subscribe(store.clientStateRefreshChan)
	store.client.OnExecRefresh().Subscribe(store.clientExecRefreshChan)
}

func (store *Store) clearClient() {
	store.clientMux.Lock()
	defer store.clientMux.Unlock()

	panics.IsTrue(store.client != nil)

	store.client.OnOnlineChanged().Unsubscribe(store.clientOnlineChangedChan)
	store.client.OnDeviceList().Unsubscribe(store.clientDeviceListChan)
	store.client.OnStateRefresh().Unsubscribe(store.clientStateRefreshChan)
	store.client.OnExecRefresh().Unsubscribe(store.clientExecRefreshChan)

	close(store.clientOnlineChangedChan)
	close(store.clientDeviceListChan)
	close(store.clientStateRefreshChan)
	close(store.clientExecRefreshChan)

	store.clientOnlineChangedChan = nil
	store.clientDeviceListChan = nil
	store.clientStateRefreshChan = nil
	store.clientExecRefreshChan = nil

	store.onOnlineChanged.Update(false)

	store.client = nil
}

func (store *Store) UnsetClient() {
	store.clearClient()

	store.mux.Lock()
	defer store.mux.Unlock()
	store.clearDevices()
}

func (store *Store) Execute(deviceURL string, command string, args ...any) {
	store.mux.Lock()
	defer store.mux.Unlock()

	if store.client == nil {
		logger.Warn("Execute without client, ignored")
		return
	}

	device := store.devices[deviceURL]
	if device == nil {
		logger.Warnf("Execute for unknown device '%s', ignored", deviceURL)
		return
	}

	store.client.Execute(&device.kiz, command, args...)
}

func (store *Store) Interrupt(deviceURL string) {
	store.mux.Lock()
	defer store.mux.Unlock()

	if store.client == nil {
		logger.Warn("Interrupt without client, ignored")
		return
	}

	device := store.devices[deviceURL]
	if device == nil {
		logger.Warnf("Interrupt for unknown device '%s', ignored", deviceURL)
		return
	}

	store.client.Interrupt(&device.kiz)
}

func (store *Store) OnOnlineChanged() tools.ObservableValue[bool] {
	return store.onOnlineChanged
}

func (store *Store) OnDeviceChanged() tools.Observable[*DeviceChange] {
	return store.onDeviceChanged
}

func (store *Store) OnStateChanged() tools.Observable[*DeviceState] {
	return store.onStateChanged
}

func (store *Store) OnExecChanged() tools.Observable[*ExecChange] {
	return store.onExecChanged
}

func (store *Store) GetDevice(deviceURL string) *Device {
	store.mux.Lock()
	defer store.mux.Unlock()

	return store.devices[deviceURL]
}

func (store *Store) GetState(deviceURL string, stateName string) *DeviceState {
	store.mux.Lock()
	defer store.mux.Unlock()

	key := store.makeStateKey(deviceURL, stateName)
	return store.states[key]
}

func (store *Store) makeStateKey(deviceURL string, stateName string) string {
	return deviceURL + "$" + stateName
}

func (store *Store) handleOnlineChanged(online bool) {
	store.onOnlineChanged.Update(online)
	// for now we consider devices stay even if offline (and states will stay accurate ...)
}

func (store *Store) handleDeviceList(devices []kizcool.Device) {
	store.mux.Lock()
	defer store.mux.Unlock()

	list := make(map[string]struct{})

	// add/update
	for _, device := range devices {
		deviceURL := string(device.DeviceURL)
		list[deviceURL] = struct{}{}
		existing := store.devices[deviceURL]

		if existing == nil {
			store.addDevice(&Device{
				kiz: device,
			})
		}

		store.refreshState(string(device.DeviceURL), device.States)
	}

	// remove
	for _, device := range maps.Values(store.devices) {
		deviceURL := device.DeviceURL()
		if _, exists := list[deviceURL]; !exists {
			store.removeDevice(device)
		}
	}
}

func (store *Store) handleStateRefresh(arg *StateRefresh) {
	store.mux.Lock()
	defer store.mux.Unlock()

	store.refreshState(arg.DeviceURL(), arg.States())
}

func (store *Store) handleExecRefresh(arg *ExecChange) {
	store.onExecChanged.Notify(arg)
}

func (store *Store) IsExecuting(deviceURL string) bool {
	store.mux.Lock()
	defer store.mux.Unlock()

	if store.client == nil {
		return false
	}

	device := store.devices[deviceURL]
	if device == nil {
		return false
	}

	return store.client.IsExecuting(&device.kiz)
}

func (store *Store) addDevice(device *Device) {
	deviceURL := device.DeviceURL()

	store.devices[deviceURL] = device

	logger.Debugf("Device joined '%s", deviceURL)

	store.onDeviceChanged.Notify(&DeviceChange{
		device:    device,
		operation: OperationAdd,
	})
}

func (store *Store) removeDevice(device *Device) {
	deviceURL := device.DeviceURL()

	delete(store.devices, deviceURL)

	logger.Debugf("Device left '%s", deviceURL)

	store.onDeviceChanged.Notify(&DeviceChange{
		device:    device,
		operation: OperationRemove,
	})
}

func (store *Store) clearDevices() {
	for _, device := range maps.Values(store.devices) {
		delete(store.devices, device.DeviceURL())
		store.onDeviceChanged.Notify(&DeviceChange{
			device:    device,
			operation: OperationRemove,
		})
	}
}

func (store *Store) refreshState(deviceURL string, states []kizcool.DeviceState) {
	for _, state := range states {
		key := store.makeStateKey(deviceURL, string(state.Name))
		oldState := store.states[key]

		if oldState == nil || !reflect.DeepEqual(oldState.value, state.Value) {
			newState := &DeviceState{
				deviceURL: deviceURL,
				name:      string(state.Name),
				typ:       state.Type,
				value:     state.Value,
			}

			store.states[key] = newState
			store.onStateChanged.Notify(newState)
		}
	}
}

type storeContainer struct {
	store    *Store
	refCount int
}

var repository = make(map[string]*storeContainer)
var repoMux sync.Mutex

func GetStore(boxKey string) *Store {
	repoMux.Lock()
	defer repoMux.Unlock()

	container, ok := repository[boxKey]
	if !ok {
		logger.Infof("Creating new store for box key '%s'", boxKey)
		container = &storeContainer{
			store:    newStore(),
			refCount: 0,
		}

		repository[boxKey] = container
	}

	container.refCount += 1
	return container.store
}

func ReleaseStore(boxKey string) {
	repoMux.Lock()
	defer repoMux.Unlock()

	container := repository[boxKey]
	panics.IsTrue(container.refCount > 0)
	container.refCount -= 1

	if container.refCount == 0 {
		logger.Infof("Removing store for box key '%s'", boxKey)
		delete(repository, boxKey)
	}
}
