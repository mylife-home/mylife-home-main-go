package bus

import (
	"context"
	"mylife-home-common/config"
	"mylife-home-common/executor"
	"mylife-home-common/tools"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type busConfig struct {
	ServerUrl string `mapstructure:"serverUrl"`
}

type OnlineChangedHandler func(bool)

type message struct {
	instanceName string
	domain       string
	path         string
	payload      []byte
	retained     bool
}

func (m *message) InstanceName() string {
	return m.instanceName
}

func (m *message) Domain() string {
	return m.domain
}

func (m *message) Path() string {
	return m.path
}

func (m *message) Payload() []byte {
	return m.payload
}

func (m *message) Retained() bool {
	return m.retained
}

type client struct {
	mqtt     mqtt.Client
	exec     executor.Executor
	ctx      context.Context // canceled on terminate
	ctxClose func()

	instanceName    string
	online          bool
	onOnlineChanged *tools.CallbackManager[bool]
	onMessage       *tools.CallbackManager[*message]
	subscriptions   map[string]struct{}
}

func newClient(instanceName string) *client {
	conf := busConfig{}
	config.BindStructure("bus", &conf)

	// Need it in advance
	client := &client{
		exec:            executor.CreateExecutor(),
		instanceName:    instanceName,
		onOnlineChanged: tools.NewCallbackManager[bool](),
		onMessage:       tools.NewCallbackManager[*message](),
		subscriptions:   make(map[string]struct{}),
	}

	client.ctx, client.ctxClose = context.WithCancel(context.Background())

	options := mqtt.NewClientOptions()
	options.AddBroker(conf.ServerUrl)
	options.SetClientID(instanceName)
	options.SetCleanSession(true)
	options.SetResumeSubs(false)
	options.SetConnectRetry(true)
	options.SetMaxReconnectInterval(time.Second * 5)
	options.SetConnectRetryInterval(time.Second * 5)

	options.SetBinaryWill(client.BuildTopic(presenceDomain), []byte{}, 0, true)

	options.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		client.exec.Execute(func() {
			client.onConnectionLost(err)
		})
	})

	options.SetOnConnectHandler(func(_ mqtt.Client) {
		client.exec.Execute(func() {
			client.onConnect()
		})
	})

	options.SetDefaultPublishHandler(func(_ mqtt.Client, m mqtt.Message) {
		client.exec.Execute(func() {
			// Investigate deadlocks (disabled because very verbose)
			// logger.Debugf("On message begin %s", m.Topic())
			// defer logger.Debugf("On message end %s", m.Topic())

			var instanceName, domain, path string
			parts := strings.SplitN(m.Topic(), "/", 3)
			count := len(parts)

			if count > 0 {
				instanceName = parts[0]
			}
			if count > 1 {
				domain = parts[1]
			}
			if count > 2 {
				path = parts[2]
			}

			client.onMessage.Execute(&message{
				instanceName: instanceName,
				domain:       domain,
				path:         path,
				payload:      m.Payload(),
				retained:     m.Retained(),
			})
		})
	})

	client.mqtt = mqtt.NewClient(options)

	// Note: with auto-retry, the connection may not fail, or for a severe reason (eg: bad config)
	client.goToken(client.mqtt.Connect())

	return client
}

func (client *client) goToken(token mqtt.Token) {
	go func() {
		token.Wait()

		if token.Error() != nil && !client.mqtt.IsConnected() {
			return
		}

		if err := token.Error(); err != nil {
			logger.WithError(err).Error("error on token")
		}
	}()
}

func (client *client) Terminate() {
	client.ctxClose()

	if client.mqtt.IsConnected() {
		client.ClearRetain(client.BuildTopic(presenceDomain))

		// TODO: should we go async?
		client.clearResidentState()
	}

	client.mqtt.Disconnect(100)
	client.exec.Terminate()
}

func (client *client) InstanceName() string {
	return client.instanceName
}

func (client *client) onConnectionLost(err error) {
	l := logger
	if err != nil {
		l = l.WithError(err)
	}

	l.Error("connection lost")

	client.onlineChanged(false)
}

func (client *client) onConnect() {
	// given the spec, it is unclear if LWT should be executed in case of client takeover, so we run it to be sure
	client.ClearRetain(client.BuildTopic(presenceDomain))

	// TODO: should we go async?
	client.clearResidentState()

	client.Publish(client.BuildTopic(presenceDomain), Encoding.WriteBool(true), true)

	client.onlineChanged(true)

	if len(client.subscriptions) > 0 {
		topics := make(map[string]byte)

		for topic := range client.subscriptions {
			topics[topic] = 0 // Topic => QoS
		}

		client.goToken(client.mqtt.SubscribeMultiple(topics, nil))
	}
}

func (client *client) OnMessage() tools.CallbackRegistration[*message] {
	return client.onMessage
}

func (client *client) OnOnlineChanged() tools.CallbackRegistration[bool] {
	return client.onOnlineChanged
}

func (client *client) onlineChanged(value bool) {
	if value == client.online {
		return
	}

	client.online = value
	logger.Infof("online: %t", value)

	client.onOnlineChanged.Execute(value)
}

func (client *client) Online() bool {
	return client.online
}

func (client *client) clearResidentState() {
	// register on self state, and remove on every message received
	// wait 1 sec after last message receive

	topicCleared := make(chan struct{})

	clearTopic := func(_ mqtt.Client, m mqtt.Message) {
		// only clear real retained messages
		if m.Retained() && len(m.Payload()) > 0 && strings.HasPrefix(m.Topic(), client.instanceName+"/") {
			client.ClearRetain(m.Topic())
			topicCleared <- struct{}{} // reset timeout
		}
	}

	selfStateTopic := client.BuildTopic("#")

	client.goToken(client.mqtt.Subscribe(selfStateTopic, 0, clearTopic))
	defer client.goToken(client.mqtt.Unsubscribe(selfStateTopic))

	for {
		select {
		case <-topicCleared:
			// reset timer on new topic

		case <-time.After(time.Second):
			// timeout, exit
			return

		case <-client.ctx.Done():
			// client exiting, exit
			return
		}
	}
}

func (client *client) BuildTopic(domain string, args ...string) string {
	finalArgs := append([]string{client.instanceName, domain}, args...)
	return strings.Join(finalArgs, "/")
}

func (client *client) BuildRemoteTopic(targetInstance string, domain string, args ...string) string {
	finalArgs := append([]string{targetInstance, domain}, args...)
	return strings.Join(finalArgs, "/")
}

func (client *client) ClearRetain(topic string) {
	client.goToken(client.mqtt.Publish(topic, 0, true, []byte{}))
}

func (client *client) Publish(topic string, payload []byte, retain bool) {
	client.goToken(client.mqtt.Publish(topic, 0, retain, payload))
}

func (client *client) Subscribe(topics ...string) {
	for _, topic := range topics {
		client.subscriptions[topic] = struct{}{}
	}

	if client.Online() {
		m := make(map[string]byte)

		for _, topic := range topics {
			m[topic] = 0 // Topic => QoS
		}

		client.goToken(client.mqtt.SubscribeMultiple(m, nil))
	}
}

func (client *client) Unsubscribe(topics ...string) {
	for _, topic := range topics {
		delete(client.subscriptions, topic)
	}

	if client.Online() {
		client.goToken(client.mqtt.Unsubscribe(topics...))
	}
}
