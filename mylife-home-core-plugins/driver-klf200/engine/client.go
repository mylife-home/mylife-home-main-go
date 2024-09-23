package engine

import (
	"context"
	"mylife-home-common/tools"
	"sync"
	"time"

	"github.com/mylife-home/klf200-go"
	"github.com/mylife-home/klf200-go/commands"
	"golang.org/x/exp/maps"
)

const timeout = time.Second * 10
const refreshDevices = time.Minute * 1
const refreshStates = time.Second * 10

type Client struct {
	address       string
	password      string
	client        *klf200.Client
	online        tools.SubjectValue[bool]
	devices       tools.Subject[[]*Device]
	deviceIndexes []int
	states        tools.Subject[[]*klf200.StatusData]

	ctx        context.Context
	close      context.CancelFunc
	workerSync sync.WaitGroup
}

func MakeClient(address string, password string) *Client {
	ctx, close := context.WithCancel(context.Background())

	client := &Client{
		address:       address,
		password:      password,
		client:        klf200.MakeClient(address, password, klf200logger),
		online:        tools.MakeSubjectValue(false),
		devices:       tools.MakeSubject[[]*Device](),
		deviceIndexes: make([]int, 0),
		states:        tools.MakeSubject[[]*klf200.StatusData](),
		ctx:           ctx,
		close:         close,
	}

	client.client.RegisterStatusChange(client.statusChange)

	client.workerSync.Add(1)
	go client.worker()

	client.client.Start()

	return client
}

func (client *Client) Terminate() {
	client.close()
	client.workerSync.Wait()

	client.client.Close()
}

func (client *Client) statusChange(cs klf200.ConnectionStatus) {
	switch cs {
	case klf200.ConnectionOpen:
		client.checkReboot()
		client.online.Update(true)
		// refresh devices/states on connection
		client.refreshDevices()
		client.refreshStates()
	case klf200.ConnectionClosed, klf200.ConnectionHandshaking:
		client.online.Update(false)
	}
}

func (client *Client) Online() tools.ObservableValue[bool] {
	return client.online
}

func (client *Client) Devices() tools.Observable[[]*Device] {
	return client.devices
}

func (client *Client) States() tools.Observable[[]*klf200.StatusData] {
	return client.states
}

func (client *Client) worker() {
	defer client.workerSync.Done()

	for {
		select {
		case <-client.ctx.Done():
			return
		case <-time.After(refreshDevices):
			client.refreshDevices()
		case <-time.After(refreshStates):
			client.refreshStates()
		}
	}
}

func (client *Client) refreshDevices() {
	if !client.online.Get() {
		return
	}

	objects, err := client.getSystemTable()
	if err != nil {
		logger.WithError(err).Error("could not get system table")
		return
	}

	nodes, err := client.getNodesInfo()
	if err != nil {
		logger.WithError(err).Error("could not get nodes info")
		return
	}

	devices := make(map[int]*Device)

	for _, object := range objects {
		index := object.SystemTableIndex
		if _, exists := devices[index]; exists {
			logger.Warnf("Device %d already exists", index)
		}

		devices[index] = &Device{
			index:   index,
			address: object.ActuatorAddress,
			typ:     object.ActuatorType,
		}
	}

	for _, node := range nodes {
		index := node.NodeID
		dev, exists := devices[index]

		if !exists {
			logger.Warnf("Device %d does not exist", index)
			continue
		}

		dev.name = node.Name
	}

	client.devices.Notify(maps.Values(devices))
	client.deviceIndexes = maps.Keys(devices)
}

func (client *Client) refreshStates() {
	if !client.online.Get() {
		return
	}

	states, err := client.getStatus(client.deviceIndexes)
	if err != nil {
		logger.WithError(err).Error("could not get states")
		return
	}

	client.states.Notify(states)
}

func (client *Client) getSystemTable() ([]commands.SystemtableObject, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxCancel()

	return client.client.Config().GetSystemTable(ctx)
}

func (client *Client) getNodesInfo() ([]*commands.GetAllNodesInformationNtf, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxCancel()

	return client.client.Info().GetAllNodesInformation(ctx)
}

func (client *Client) getStatus(nodeIndexes []int) ([]*klf200.StatusData, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxCancel()

	return client.client.Commands().Status(ctx, nodeIndexes)
}

func (client *Client) Reboot() {
	if err := client.client.Device().Reboot(); err != nil {
		logger.WithError(err).Error("could not reboot")
	}

	logger.Info("Box rebooted")
}

func (client *Client) ChangePosition(nodeIndex int, position commands.MPValue) (*klf200.Session, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxCancel()

	return client.client.Commands().ChangePosition(ctx, nodeIndex, position)
}

func (client *Client) Mode(nodeIndex int) (*klf200.Session, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxCancel()

	return client.client.Commands().Mode(ctx, nodeIndex)
}

// KLF200 has 2 possible sockets (connections). But if a connection is hardly broken, KLF200 leaks it.
// On connection, we check if we can do another connection, and properly disconnect.
// If we cannot, the other connection is leaked, and we reboot.
func (client *Client) checkReboot() {
	if !client.checkSecondConnection() {
		logger.Info("Could not use the second box connection, rebooting")
		client.Reboot()
	}
}

func (client *Client) checkSecondConnection() bool {
	connOk := make(chan struct{})

	testClient := klf200.MakeClient(client.address, client.password, klf200logger)
	testClient.RegisterStatusChange(func(cs klf200.ConnectionStatus) {
		if cs == klf200.ConnectionOpen {
			close(connOk)
		}
	})

	testClient.Start()
	defer testClient.Close()

	select {
	case <-time.After(time.Second * 30):
		return false
	case <-connOk:
		return true
	}
}
