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
}

func NewTransport(options *Options) *Transport {
	client := newClient(defines.InstanceName())
	transport := &Transport{
		client:     client,
		rpc:        newRpc(client),
		presence:   newPresence(client, options.presenceTracking),
		components: newComponents(client),
		metadata:   newMetadata(client),
		logger:     newLogger(client),
	}

	transport.client.OnOnlineChanged().Register(func(online bool) {
		if online {
			transport.publishInstanceInfo()
		}
	})

	instance_info.OnUpdate().Register(func(_ *instance_info.InstanceInfo) {
		if transport.Online() {
			transport.publishInstanceInfo()
		}
	})

	return transport
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

func (transport *Transport) OnOnlineChanged() tools.CallbackRegistration[bool] {
	return transport.client.OnOnlineChanged()
}

func (transport *Transport) Online() bool {
	return transport.client.Online()
}

func (transport *Transport) publishInstanceInfo() {
	data := instance_info.Get()
	transport.metadata.Set("instance-info", data)
}
