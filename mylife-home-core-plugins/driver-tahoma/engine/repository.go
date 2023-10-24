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
	client                   *Client
	clientOnlineChangedToken tools.RegistrationToken
	clientDeviceListToken    tools.RegistrationToken
	clientStateRefreshToken  tools.RegistrationToken
	clientExecRefreshToken   tools.RegistrationToken

	devices         map[string]*Device
	states          map[string]*DeviceState // key = <deviceURL>$<name>
	onOnlineChanged *tools.CallbackManager[bool]
	onDeviceChanged *tools.CallbackManager[*DeviceChange]
	onStateChanged  *tools.CallbackManager[*DeviceState]
	onExecChanged   *tools.CallbackManager[*ExecChange]

	mux sync.Mutex
}

func newStore() *Store {
	return &Store{
		devices:         make(map[string]*Device),
		states:          make(map[string]*DeviceState),
		onOnlineChanged: tools.NewCallbackManager[bool](),
		onDeviceChanged: tools.NewCallbackManager[*DeviceChange](),
		onStateChanged:  tools.NewCallbackManager[*DeviceState](),
		onExecChanged:   tools.NewCallbackManager[*ExecChange](),
	}
}

func (store *Store) SetClient(client *Client) {
	store.mux.Lock()
	defer store.mux.Unlock()

	store.client = client
	store.clientOnlineChangedToken = store.client.OnOnlineChanged().Register(store.handleOnlineChanged)
	store.clientDeviceListToken = store.client.OnDeviceList().Register(store.handleDeviceList)
	store.clientStateRefreshToken = store.client.OnStateRefresh().Register(store.handleStateRefresh)
	store.clientExecRefreshToken = store.client.OnExecRefresh().Register(store.handleExecRefresh)

	if client.Online() {
		store.onOnlineChanged.Execute(true)
	}
}

func (store *Store) UnsetClient() {
	store.mux.Lock()
	defer store.mux.Unlock()

	if store.client.Online() {
		store.onOnlineChanged.Execute(false)

		for _, device := range maps.Values(store.devices) {
			delete(store.devices, device.DeviceURL())
			store.onDeviceChanged.Execute(&DeviceChange{
				device:    device,
				operation: OperationRemove,
			})
		}

		clear(store.states)
	}

	store.client.OnOnlineChanged().Unregister(store.clientOnlineChangedToken)
	store.client.OnDeviceList().Unregister(store.clientDeviceListToken)
	store.client.OnStateRefresh().Unregister(store.clientStateRefreshToken)
	store.client.OnExecRefresh().Unregister(store.clientExecRefreshToken)

	store.clientOnlineChangedToken = tools.InvalidRegistrationToken
	store.clientDeviceListToken = tools.InvalidRegistrationToken
	store.clientStateRefreshToken = tools.InvalidRegistrationToken
	store.clientExecRefreshToken = tools.InvalidRegistrationToken
	store.client = nil
}

func (store *Store) Execute(deviceURL string, command string, args ...[]any) {
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

func (store *Store) OnOnlineChanged() tools.CallbackRegistration[bool] {
	return store.onOnlineChanged
}

func (store *Store) OnDeviceChanged() tools.CallbackRegistration[*DeviceChange] {
	return store.onDeviceChanged
}

func (store *Store) OnStateChanged() tools.CallbackRegistration[*DeviceState] {
	return store.onStateChanged
}

func (store *Store) OnExecChanged() tools.CallbackRegistration[*ExecChange] {
	return store.onExecChanged
}

func (store *Store) Online() bool {
	store.mux.Lock()
	defer store.mux.Unlock()

	return store.client != nil && store.client.Online()
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
	store.onOnlineChanged.Execute(online)
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
			newDevice := &Device{
				kiz: device,
			}

			store.devices[deviceURL] = newDevice
			store.onDeviceChanged.Execute(&DeviceChange{
				device:    newDevice,
				operation: OperationAdd,
			})
		}

		store.handleStateRefresh(&StateRefresh{
			deviceURL: string(device.DeviceURL),
			states:    device.States,
		})
	}

	// remove
	for _, device := range maps.Values(store.devices) {
		deviceURL := device.DeviceURL()
		if _, exists := list[deviceURL]; !exists {
			delete(store.devices, deviceURL)
			store.onDeviceChanged.Execute(&DeviceChange{
				device:    device,
				operation: OperationRemove,
			})
		}
	}
}

func (store *Store) handleStateRefresh(arg *StateRefresh) {
	store.mux.Lock()
	defer store.mux.Unlock()

	for _, state := range arg.States() {
		key := store.makeStateKey(arg.DeviceURL(), string(state.Name))
		oldState := store.states[key]

		if oldState == nil || !reflect.DeepEqual(oldState.value, state.Value) {
			newState := &DeviceState{
				deviceURL: arg.DeviceURL(),
				name:      string(state.Name),
				typ:       state.Type,
				value:     state.Value,
			}

			store.states[key] = newState
			store.onStateChanged.Execute(newState)
		}
	}
}

func (store *Store) handleExecRefresh(arg *ExecChange) {
	store.onExecChanged.Execute(arg)
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
