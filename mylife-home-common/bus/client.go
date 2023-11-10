package bus

import (
	"errors"
	"mylife-home-common/config"
	"mylife-home-common/tools"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"golang.org/x/sync/errgroup"
)

var errClosing = errors.New("client closing")

type busConfig struct {
	ServerUrl string `mapstructure:"serverUrl"`
}

type OnlineChangedHandler func(bool)

type message struct {
	topic        string
	instanceName string
	domain       string
	path         string
	payload      []byte
	retained     bool
}

func (m *message) Topic() string {
	return m.topic
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
	instanceName string
	mqtt         mqtt.Client
	online       tools.SubjectValue[bool]
}

func newClient(instanceName string) *client {
	conf := busConfig{}
	config.BindStructure("bus", &conf)

	// Need it in advance
	client := &client{
		instanceName: instanceName,
		online:       tools.MakeSubjectValue[bool](false),
	}

	options := mqtt.NewClientOptions()
	options.AddBroker(conf.ServerUrl)
	options.SetClientID(instanceName)
	options.SetCleanSession(true)
	options.SetResumeSubs(false)
	options.SetConnectRetry(true)
	options.SetMaxReconnectInterval(time.Second * 5)
	options.SetConnectRetryInterval(time.Second * 5)
	options.SetOrderMatters(false)
	options.SetConnectRetry(false)
	options.SetOrderMatters(true)

	options.SetBinaryWill(client.BuildTopic(presenceDomain), []byte{}, 0, true)

	options.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		logger.WithError(err).Error("Connection lost")
		client.online.Update(false)
	})

	options.SetOnConnectHandler(func(_ mqtt.Client) {
		client.onConnect()
	})

	client.mqtt = mqtt.NewClient(options)

	// Note: with auto-retry, the connection may not fail, or for a severe reason (eg: bad config)
	token := client.mqtt.Connect()
	go func() {
		client.wait(token)
	}()

	return client
}

func (client *client) Terminate() {
	if client.mqtt.IsConnectionOpen() {
		if err := client.ClearRetain(client.BuildTopic(presenceDomain)); err != nil {
			logger.WithError(err).Errorf("Error clearing presence")
		}

		if err := client.clearResidentState(); err != nil {
			logger.WithError(err).Errorf("Error cleaning resident state")
		}
	}

	client.mqtt.Disconnect(100)
}

func (client *client) InstanceName() string {
	return client.instanceName
}

func (client *client) onConnect() {
	go func() {
		if err := client.clearResidentState(); err != nil {
			if !client.mqtt.IsConnectionOpen() {
				return
			}

			logger.WithError(err).Errorf("Error cleaning resident state")
			return
			// TODO: should we retry?
		}

		if err := client.PublishRetain(client.BuildTopic(presenceDomain), Encoding.WriteBool(true)); err != nil {
			if !client.mqtt.IsConnectionOpen() {
				return
			}

			logger.WithError(err).Errorf("Error cleaning resident state")
			return
			// TODO: should we retry?
		}

		logger.Info("Connection established")
		client.online.Update(true)
	}()
}

func (client *client) clearResidentState() error {
	// register on self state, and remove on every message received
	// wait 1 sec after last message receive

	topicQueue := make(chan string, 1024)

	onTopic := func(m *message) {
		// only clear real retained messages
		if m.Retained() && len(m.Payload()) > 0 && m.InstanceName() == client.instanceName {
			topicQueue <- m.Topic()
		}
	}

	selfStateTopic := client.BuildTopic("#")

	if err := client.Subscribe(selfStateTopic, onTopic); err != nil {
		return err
	}

	exitLoop := false

	errg := errgroup.Group{}

	for !exitLoop {
		select {
		case topic := <-topicQueue:
			logger.Debugf("Clear resident state: '%s'", topic)

			// reset timer on new topic + actually clear it
			errg.Go(func() error {
				return client.ClearRetain(topic)
			})

		case <-time.After(time.Second):
			// timeout, exit
			exitLoop = true
		}
	}

	if err := client.Unsubscribe(selfStateTopic); err != nil {
		return err
	}

	if err := errg.Wait(); err != nil {
		return err
	}

	return nil
}

func (client *client) Online() tools.ObservableValue[bool] {
	return client.online
}

func (client *client) BuildTopic(domain string, args ...string) string {
	return client.BuildRemoteTopic(client.instanceName, domain, args...)
}

func (client *client) BuildRemoteTopic(targetInstance string, domain string, args ...string) string {
	finalArgs := append([]string{targetInstance, domain}, args...)
	return strings.Join(finalArgs, "/")
}

func (client *client) ClearRetain(topic string) error {
	return client.PublishRetain(topic, []byte{})
}

func (client *client) PublishRetain(topic string, payload []byte) error {
	return client.wait(client.mqtt.Publish(topic, 0, true, payload))
}

func (client *client) Publish(topic string, payload []byte) error {
	return client.wait(client.mqtt.Publish(topic, 0, false, payload))
}

func (client *client) Subscribe(topic string, callback func(m *message)) error {

	cb := func(_ mqtt.Client, m mqtt.Message) {
		// logger.Debugf("Got message %s", m.Topic())

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

		callback(&message{
			topic:        m.Topic(),
			instanceName: instanceName,
			domain:       domain,
			path:         path,
			payload:      m.Payload(),
			retained:     m.Retained(),
		})
	}

	return client.wait(client.mqtt.Subscribe(topic, 0, cb))
}

func (client *client) Unsubscribe(topics ...string) error {
	if !client.mqtt.IsConnectionOpen() {
		return nil
	}

	return client.wait(client.mqtt.Unsubscribe(topics...))
}

func (client *client) wait(token mqtt.Token) error {
	token.Wait()
	return token.Error()
}
