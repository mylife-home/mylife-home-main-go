package bus

import (
	"mylife-home-common/defines"
	"mylife-home-common/instance_info"
	"mylife-home-common/log"
	"mylife-home-common/tools"
	"sync"
)

var logger = log.CreateLogger("mylife:home:bus")

type Transport struct {
	client     *client
	rpc        *Rpc
	presence   *Presence
	components *Components
	metadata   *Metadata
	logger     *Logger

	onlineChan             chan bool
	instanceInfoUpdateChan chan *instance_info.InstanceInfo
}

func NewTransport() *Transport {
	client := newClient(defines.InstanceName())

	transport := &Transport{
		client: client,

		rpc:        newRpc(client),
		presence:   newPresence(client),
		components: newComponents(client),
		metadata:   newMetadata(client),
		logger:     newLogger(client),

		onlineChan:             make(chan bool, 10),
		instanceInfoUpdateChan: make(chan *instance_info.InstanceInfo, 10),
	}

	go transport.publishWorker()

	transport.client.Online().Subscribe(transport.onlineChan)
	instance_info.OnUpdate().Subscribe(transport.instanceInfoUpdateChan)

	return transport
}

func (transport *Transport) Terminate() {
	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		transport.client.Online().Unsubscribe(transport.onlineChan)
		instance_info.OnUpdate().Unsubscribe(transport.instanceInfoUpdateChan)
		close(transport.onlineChan)
		close(transport.instanceInfoUpdateChan)
	}()

	go func() {
		defer wg.Done()
		transport.presence.terminate()
	}()

	go func() {
		defer wg.Done()
		transport.rpc.terminate()
	}()

	wg.Wait()

	transport.client.Terminate()
}

func (transport *Transport) publishWorker() {
	onlineChan := transport.onlineChan
	instanceInfoUpdateChan := transport.instanceInfoUpdateChan

	for onlineChan != nil || instanceInfoUpdateChan != nil {
		var instanceInfo *instance_info.InstanceInfo

		select {
		case online, ok := <-onlineChan:
			if !ok {
				onlineChan = nil // closing
				continue
			}

			if !online {
				continue
			}

			instanceInfo = instance_info.Get()

		case data, ok := <-instanceInfoUpdateChan:
			if !ok {
				instanceInfoUpdateChan = nil // closing
				continue
			}

			if !transport.client.Online().Get() {
				continue
			}

			instanceInfo = data
		}

		switch err := transport.metadata.Set("instance-info", instanceInfo); {
		case err == nil, err == errClosing:
			// OK
		default:
			logger.WithError(err).Error("could not publish instance info")
		}
	}
}

func (transport *Transport) Rpc() *Rpc {
	return transport.rpc
}

func (transport *Transport) Presence() *Presence {
	return transport.presence
}

// TODO
func (transport *Transport) Components() *Components {
	return transport.components
}

func (transport *Transport) Metadata() *Metadata {
	return transport.metadata
}

// TODO
func (transport *Transport) Logger() *Logger {
	return transport.logger
}

func (transport *Transport) Online() tools.ObservableValue[bool] {
	return transport.client.Online()
}
