package bus

import (
	"mylife-home-common/defines"
	"mylife-home-common/instance_info"
	"mylife-home-common/log"
	"mylife-home-common/tools"
)

var logger = log.CreateLogger("mylife:home:bus")

type Options struct {
	presenceTracking bool
}

func (options *Options) SetPresenceTracking(value bool) *Options {
	options.presenceTracking = value
	return options
}

func NewOptions() *Options {
	return &Options{
		presenceTracking: false,
	}
}

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

func NewTransport(options *Options) *Transport {
	client := newClient(defines.InstanceName())
	transport := &Transport{
		client:                 client,
		rpc:                    newRpc(client),
		presence:               newPresence(client, options.presenceTracking),
		components:             newComponents(client),
		metadata:               newMetadata(client),
		logger:                 newLogger(client),
		onlineChan:             make(chan bool, 10),
		instanceInfoUpdateChan: make(chan *instance_info.InstanceInfo, 10),
	}

	go transport.publishWorker()

	transport.client.online.Subscribe(transport.onlineChan)
	instance_info.OnUpdate().Subscribe(transport.instanceInfoUpdateChan)

	return transport
}

func (transport *Transport) publishWorker() {
	for {
		var instanceInfo *instance_info.InstanceInfo

		select {
		case online, ok := <-transport.onlineChan:
			if !ok {
				return // closing
			}

			if online {
				instanceInfo = instance_info.Get()
			}

		case data, ok := <-transport.instanceInfoUpdateChan:
			if !ok {
				return // closing
			}

			if transport.client.Online().Get() {
				instanceInfo = data
			}
		}

		switch err := transport.metadata.Set("instance-info", instanceInfo); {
		case err == nil, err == errClosing:
			// OK
		default:
			logger.WithError(err).Error("could not publish instance info")
		}
	}
}

func (transport *Transport) Terminate() {
	transport.client.online.Unsubscribe(transport.onlineChan)
	instance_info.OnUpdate().Unsubscribe(transport.instanceInfoUpdateChan)
	close(transport.onlineChan)
	close(transport.instanceInfoUpdateChan)

	transport.client.Terminate()
}

func (transport *Transport) Rpc() *Rpc {
	return transport.rpc
}

func (transport *Transport) Presence() *Presence {
	return transport.presence
}

func (transport *Transport) Components() *Components {
	return transport.components
}

func (transport *Transport) Metadata() *Metadata {
	return transport.metadata
}

func (transport *Transport) Logger() *Logger {
	return transport.logger
}

func (transport *Transport) Online() tools.ObservableValue[bool] {
	return transport.client.Online()
}
