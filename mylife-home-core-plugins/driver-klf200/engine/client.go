package engine

import (
	"context"
	"mylife-home-common/tools"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/mylife-home/klf200-go"
	"github.com/mylife-home/klf200-go/commands"
	"golang.org/x/exp/maps"
)

const timeout = time.Second * 10
const refreshDevices = time.Minute * 1

type Client struct {
	client  *klf200.Client
	online  tools.SubjectValue[bool]
	devices tools.Subject[[]*Device]

	ctx        context.Context
	close      context.CancelFunc
	workerSync sync.WaitGroup
}

func MakeClient(address string, password string) *Client {
	ctx, close := context.WithCancel(context.Background())

	client := &Client{
		client:  klf200.MakeClient(address, password, klf200logger),
		online:  tools.MakeSubjectValue(false),
		devices: tools.MakeSubject[[]*Device](),
		ctx:     ctx,
		close:   close,
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
		client.online.Update(true)
		// refresh devices on connection
		client.refreshDevices()
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

func (client *Client) worker() {
	defer client.workerSync.Done()

	for {
		select {
		case <-client.ctx.Done():
			return
		case <-time.After(refreshDevices):
			client.refreshDevices()
		}
	}

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

func (client *Client) refreshDevices() {
	if !client.online.Get() {
		return
	}

	objects, err := client.getSystemTable()
	if err != nil {
		log.WithError(err).Error("could not get system table")
		return
	}

	nodes, err := client.getNodesInfo()
	if err != nil {
		log.WithError(err).Error("could not get nodes info")
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
}
