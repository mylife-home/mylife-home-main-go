package engine

import (
	"context"
	"mylife-home-common/tools"
	"sync"
	"time"

	"github.com/sgrimee/kizcool"
)

type StateRefresh struct {
	deviceURL string
	states    []kizcool.DeviceState
}

func (arg *StateRefresh) DeviceURL() string {
	return arg.deviceURL
}

func (arg *StateRefresh) States() []kizcool.DeviceState {
	return arg.states
}

type Client struct {
	kiz            *kizcool.Kiz
	workerContext  context.Context
	workerClose    func()
	workedExited   chan struct{}
	triggerRefresh func()
	executions     *runningExecutions
	online         tools.SubjectValue[bool]
	onDeviceList   tools.Subject[[]kizcool.Device]
	onStateRefresh tools.Subject[*StateRefresh]
	onExecRefresh  tools.Subject[*ExecChange]
}

const tahomaUrlBase = "https://ha101-1.overkiz.com/enduser-mobile-web"
const refreshInterval = time.Minute
const pollInterval = time.Second * 2
const devicesRefreshInterval = time.Minute * 5

func MakeClient(user string, pass string) (*Client, error) {
	kiz, err := kizcool.New(user, pass, tahomaUrlBase, "")
	if err != nil {
		return nil, err
	}

	ctx, close := context.WithCancel(context.Background())

	client := &Client{
		kiz:            kiz,
		workerContext:  ctx,
		workerClose:    close,
		workedExited:   make(chan struct{}),
		online:         tools.MakeSubjectValue[bool](false),
		onDeviceList:   tools.MakeSubject[[]kizcool.Device](),
		onStateRefresh: tools.MakeSubject[*StateRefresh](),
		onExecRefresh:  tools.MakeSubject[*ExecChange](),
	}

	client.executions = newRunningExecutions(client.onExecChanged)

	go client.worker()

	return client, nil
}

func (client *Client) Terminate() {
	client.workerClose()
	<-client.workedExited
	client.kiz = nil
}

func (client *Client) Online() tools.ObservableValue[bool] {
	return client.online
}

func (client *Client) OnDeviceList() tools.Observable[[]kizcool.Device] {
	return client.onDeviceList
}

func (client *Client) OnStateRefresh() tools.Observable[*StateRefresh] {
	return client.onStateRefresh
}

func (client *Client) OnExecRefresh() tools.Observable[*ExecChange] {
	return client.onExecRefresh
}

func (client *Client) setConnected(value bool) {
	if client.online.Update(value) {
		logger.Infof("Connected = %t", value)
	}
}

func (client *Client) onExecChanged(deviceURL kizcool.DeviceURL, executing bool) {
	client.onExecRefresh.Notify(&ExecChange{
		deviceURL: string(deviceURL),
		executing: executing,
	})
}

func (client *Client) Execute(device *kizcool.Device, command string, args ...any) {
	if !client.online.Get() {
		logger.Warnf("Client is offline, cannot run command '%s' on device '%s'.", command, device.DeviceURL)
		return
	}

	go func() {
		client.cancel(device)

		cmd := kizcool.Command{
			Name: command, Parameters: args,
		}

		action, err := kizcool.ActionGroupWithOneCommand(*device, cmd)
		if err != nil {
			logger.WithError(err).Errorf("Error at creation action for command '%s' on device '%s'", command, device.DeviceURL)
			return
		}

		execId, err := client.kiz.Execute(action)
		client.afterReq(err)
		if err != nil {
			logger.WithError(err).Errorf("Error at executing command '%s' on device '%s'", command, device.DeviceURL)
			return
		}

		client.executions.Set(device.DeviceURL, execId)
		logger.Debugf("Started execution '%s' of command '%s' on device '%s'", execId, command, device.DeviceURL)
	}()
}

func (client *Client) Interrupt(device *kizcool.Device) {
	if !client.online.Get() {
		logger.Warnf("Client is offline, cannot interrupt on device '%s'.", device.DeviceURL)
		return
	}

	go func() {
		client.cancel(device)
	}()
}

func (client *Client) cancel(device *kizcool.Device) {
	oldExecId, ok := client.executions.GetByDevice(device.DeviceURL)
	if ok {
		logger.Debugf("Canceling execution '%s'", oldExecId)

		// request('DELETE', `/exec/current/setup/${execId}`);
		_, err := client.kiz.Stop(*device)
		client.afterReq(err)
		if err != nil {
			logger.WithError(err).Errorf("Error at canceling execution '%s' of device '%s'", oldExecId, device.DeviceURL)
		}

		client.executions.RemoveByDevice(device.DeviceURL)
	}
}

func (client *Client) IsExecuting(device *kizcool.Device) bool {
	_, ok := client.executions.GetByDevice(device.DeviceURL)
	return ok
}

func (client *Client) worker() {
	defer close(client.workedExited)

	refreshTimer := time.NewTicker(refreshInterval)
	pollTimer := time.NewTicker(pollInterval)
	devicesRefreshTimer := time.NewTicker(devicesRefreshInterval)

	defer refreshTimer.Stop()
	defer pollTimer.Stop()
	defer devicesRefreshTimer.Stop()

	client.triggerRefresh = func() {
		// Trigger nearly-immediately
		refreshTimer.Reset(time.Millisecond * 100)
	}

	defer func() {
		client.triggerRefresh = nil
	}()

	client.devicesRefresh()
	client.refresh()

	for {
		select {
		case <-devicesRefreshTimer.C:
			client.devicesRefresh()

		case <-refreshTimer.C:
			refreshTimer.Reset(refreshInterval)
			client.refresh()

		case <-pollTimer.C:
			client.poll()

		case <-client.workerContext.Done():
			return
		}
	}
}

func (client *Client) afterReq(err error) {
	// consider after an error we are disconnected and after a success we are connected
	client.setConnected(err == nil)
}

func (client *Client) refresh() {
	// logger.Debug("Refresh")
	err := client.kiz.RefreshStates()
	client.afterReq(err)

	if err != nil {
		logger.WithError(err).Error("States refresh error")
	}
}

func (client *Client) poll() {
	// logger.Debug("Poll")
	events, err := client.kiz.PollEvents()
	client.afterReq(err)

	if err != nil {
		logger.WithError(err).Error("Poll error")
		return
	}

	for _, event := range events {
		client.processEvent(event)
	}
}

func (client *Client) processEvent(event kizcool.Event) {
	// logger.Debugf("Got event %+v", event)
	switch event := event.(type) {
	case *kizcool.ExecutionStateChangedEvent:
		if event.TimeToNextState == -1 {
			client.executions.RemoveByExec(event.ExecID)
			logger.Debugf("Execution ended '%s' %s", event.ExecID, event.NewState)
		}

	case *kizcool.DeviceStateChangedEvent:
		client.onStateRefresh.Notify(&StateRefresh{
			deviceURL: string(event.DeviceURL),
			states:    event.DeviceStates,
		})

	case *kizcool.RefreshAllDevicesStatesCompletedEvent:
		// Trigger refresh state immediatly after we ended prev one
		client.triggerRefresh()

	// Event below does not need special action
	case *kizcool.GatewaySynchronizationStartedEvent:
	// sent to mark a group of refresh device events
	case *kizcool.GatewaySynchronizationEndedEvent:
	case *kizcool.ExecutionRegisteredEvent:
		// sent right after execute (before execution states changes)
	case *kizcool.CommandExecutionStateChangedEvent:
		// sent after 'cancel', we already removed the execution and we will get 'ExecutionStateChangedEvent' anyway

	default:
		logger.Debugf("Unhandled event %+v", event)
	}
}

func (client *Client) devicesRefresh() {
	// logger.Debug("DevicesRefresh")
	devices, err := client.kiz.GetDevices()
	client.afterReq(err)

	if err != nil {
		logger.WithError(err).Error("Device refresh error")
		return
	}

	client.onDeviceList.Notify(devices)
}

type runningExecutions struct {
	onChange func(kizcool.DeviceURL, bool)
	byDevice map[kizcool.DeviceURL]kizcool.ExecID
	byExec   map[kizcool.ExecID]kizcool.DeviceURL
	mux      sync.Mutex
}

func newRunningExecutions(onChange func(kizcool.DeviceURL, bool)) *runningExecutions {
	return &runningExecutions{
		onChange: onChange,
		byDevice: make(map[kizcool.DeviceURL]kizcool.ExecID),
		byExec:   make(map[kizcool.ExecID]kizcool.DeviceURL),
	}
}

func (executions *runningExecutions) Clear() {
	executions.mux.Lock()
	defer executions.mux.Unlock()

	for deviceURL := range executions.byDevice {
		executions.onChange(deviceURL, false)
	}

	clear(executions.byDevice)
	clear(executions.byExec)
}

func (executions *runningExecutions) GetByDevice(deviceURL kizcool.DeviceURL) (kizcool.ExecID, bool) {
	executions.mux.Lock()
	defer executions.mux.Unlock()

	execId, ok := executions.byDevice[deviceURL]
	return execId, ok
}

func (executions *runningExecutions) Set(deviceURL kizcool.DeviceURL, execId kizcool.ExecID) {
	executions.mux.Lock()
	defer executions.mux.Unlock()

	executions.byDevice[deviceURL] = execId
	executions.byExec[execId] = deviceURL
	executions.onChange(deviceURL, true)
}

func (executions *runningExecutions) RemoveByDevice(deviceURL kizcool.DeviceURL) {
	executions.mux.Lock()
	defer executions.mux.Unlock()

	execId, ok := executions.byDevice[deviceURL]
	if ok {
		delete(executions.byDevice, deviceURL)
		delete(executions.byExec, execId)
		executions.onChange(deviceURL, false)
	}
}

func (executions *runningExecutions) RemoveByExec(execId kizcool.ExecID) {
	executions.mux.Lock()
	defer executions.mux.Unlock()

	deviceURL, ok := executions.byExec[execId]
	if ok {
		delete(executions.byDevice, deviceURL)
		delete(executions.byExec, execId)
		executions.onChange(deviceURL, false)
	}
}
