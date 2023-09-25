package bus

import (
	"mylife-home-common/config"
	"mylife-home-common/tools"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type busConfig struct {
	ServerUrl string `mapstructure:"serverUrl"`
}

type OnlineChangedHandler func(bool)
type messageHandler func(*message)

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
	instanceName      string
	mqtt              mqtt.Client
	online            bool
	onlineSync        sync.RWMutex
	onOnlineChanged   *tools.CallbackManager[bool]
	onMessage         *tools.CallbackManager[*message]
	subscriptions     map[string]struct{}
	subscriptionsSync sync.RWMutex
}

func newClient(instanceName string) *client {
	conf := busConfig{}
	config.BindStructure("bus", &conf)

	// Need it in advance
	client := &client{
		instanceName:    instanceName,
		onOnlineChanged: tools.NewCallbackManager[bool](),
		onMessage:       tools.NewCallbackManager[*message](),
		subscriptions:   make(map[string]struct{}),
	}

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
		client.onConnectionLost(err)
	})

	options.SetOnConnectHandler(func(_ mqtt.Client) {
		client.onConnect()
	})

	options.SetDefaultPublishHandler(func(_ mqtt.Client, m mqtt.Message) {
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

	client.mqtt = mqtt.NewClient(options)

	// Note: with auto-retry, the connection may not fail, or for a severe reason (eg: bad config)
	fireAndForget(func() error {
		return tokenSync(client.mqtt.Connect())
	})

	return client
}

func (client *client) Terminate() {
	if client.mqtt.IsConnected() {
		if err := client.ClearRetain(client.BuildTopic(presenceDomain)); err != nil {
			logger.WithError(err).Error("Send offline error")
		}

		if err := client.clearResidentState(); err != nil {
			logger.WithError(err).Error("Clear resident state error")
		}
	}

	client.mqtt.Disconnect(100)
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
	fireAndForget(func() error {
		// given the spec, it is unclear if LWT should be executed in case of client takeover, so we run it to be sure
		if err := client.ClearRetain(client.BuildTopic(presenceDomain)); err != nil {
			return err
		}

		if err := client.clearResidentState(); err != nil {
			return err
		}

		if err := client.Publish(client.BuildTopic(presenceDomain), Encoding.WriteBool(true), true); err != nil {
			return err
		}

		client.onlineChanged(true)

		topics := client.prepareResubscribe()
		if len(topics) > 0 {
			if err := tokenSync(client.mqtt.SubscribeMultiple(topics, nil)); err != nil {
				return err
			}
		}

		return nil
	})
}

func (client *client) prepareResubscribe() map[string]byte {
	m := make(map[string]byte)

	client.subscriptionsSync.RLock()
	defer client.subscriptionsSync.RUnlock()

	for topic := range client.subscriptions {
		m[topic] = 0 // Topic => QoS
	}

	return m
}

func (client *client) OnMessage() tools.CallbackRegistration[*message] {
	return client.onMessage
}

func (client *client) OnOnlineChanged() tools.CallbackRegistration[bool] {
	return client.onOnlineChanged
}

func (client *client) onlineChanged(value bool) {
	client.onlineSync.Lock()
	defer client.onlineSync.Unlock()

	if value == client.online {
		return
	}

	client.online = value
	logger.Infof("online: %t", value)

	client.onOnlineChanged.Execute(value)
}

func (client *client) Online() bool {
	client.onlineSync.RLock()
	defer client.onlineSync.RUnlock()

	return client.online
}

func (client *client) clearResidentState() error {
	// register on self state, and remove on every message received
	// wait 1 sec after last message receive

	newTopic := make(chan struct{}, 100)

	clearTopic := func(_ mqtt.Client, m mqtt.Message) {
		// only clear real retained messages
		if m.Retained() && len(m.Payload()) > 0 && strings.HasPrefix(m.Topic(), client.instanceName+"/") {
			newTopic <- struct{}{}
			fireAndForget(func() error {
				return client.ClearRetain(m.Topic())
			})
		}
	}

	selfStateTopic := client.BuildTopic("#")

	if err := tokenSync(client.mqtt.Subscribe(selfStateTopic, 0, clearTopic)); err != nil {
		return err
	}

	timeout := false
	for !timeout {
		select {
		case <-newTopic:
			// reset timer on new topic

		case <-time.After(time.Second):
			// timeout, exit
			timeout = true
		}
	}

	if err := tokenSync(client.mqtt.Unsubscribe(selfStateTopic)); err != nil {
		return err
	}

	return nil
}

func (client *client) BuildTopic(domain string, args ...string) string {
	finalArgs := append([]string{client.instanceName, domain}, args...)
	return strings.Join(finalArgs, "/")
}

func (client *client) BuildRemoteTopic(targetInstance string, domain string, args ...string) string {
	finalArgs := append([]string{targetInstance, domain}, args...)
	return strings.Join(finalArgs, "/")
}

func (client *client) ClearRetainAsync(topic string) mqtt.Token {
	return client.mqtt.Publish(topic, 0, true, []byte{})
}

func (client *client) ClearRetain(topic string) error {
	return tokenSync(client.ClearRetainAsync(topic))
}

func (client *client) PublishAsync(topic string, payload []byte, retain bool) mqtt.Token {
	return client.mqtt.Publish(topic, 0, retain, payload)
}

func (client *client) Publish(topic string, payload []byte, retain bool) error {
	return tokenSync(client.PublishAsync(topic, payload, retain))
}

func (client *client) SubscribeAsync(topics ...string) mqtt.Token {
	client.subscriptionsAdd(topics)

	if client.Online() {
		m := make(map[string]byte)

		for _, topic := range topics {
			m[topic] = 0 // Topic => QoS
		}

		return client.mqtt.SubscribeMultiple(m, nil)
	}

	return newDoneToken()
}

func (client *client) Subscribe(topics ...string) error {
	return tokenSync(client.SubscribeAsync(topics...))
}

func (client *client) subscriptionsAdd(topics []string) {
	client.subscriptionsSync.Lock()
	defer client.subscriptionsSync.Unlock()

	for _, topic := range topics {
		client.subscriptions[topic] = struct{}{}
	}
}

func (client *client) UnsubscribeAsync(topics ...string) mqtt.Token {
	client.subscriptionsDel(topics)

	if client.Online() {
		return client.mqtt.Unsubscribe(topics...)
	}

	return newDoneToken()
}

func (client *client) Unsubscribe(topics ...string) error {
	return tokenSync(client.UnsubscribeAsync(topics...))
}

func (client *client) subscriptionsDel(topics []string) {
	client.subscriptionsSync.Lock()
	defer client.subscriptionsSync.Unlock()

	for _, topic := range topics {
		delete(client.subscriptions, topic)
	}
}

func fireAndForget(callback func() error) {
	go func() {
		if err := callback(); err != nil {
			logger.WithError(err).Error("Fire and forget failed")
		}
	}()
}

func tokenSync(token mqtt.Token) error {
	token.Wait()
	return token.Error()
}
