package bus

import (
	"fmt"
	"mylife-home-common/tools"
	"strings"
)

const componentsDomain = "components"

type LocalComponent interface {
	RegisterAction(name string, handler func([]byte))
	SetState(name string, value []byte)
}

type RemoteComponent interface {
	EmitAction(name string, value []byte)
	RegisterStateChange(name string, handler func([]byte))
}

type Components struct {
	client           *client
	localComponents  map[string]*localComponentImpl
	remoteComponents map[*remoteComponentImpl]struct{}
}

func newComponents(client *client) *Components {
	return &Components{
		client:           client,
		localComponents:  make(map[string]*localComponentImpl),
		remoteComponents: make(map[*remoteComponentImpl]struct{}),
	}
}

func (comps *Components) AddLocalComponent(id string) (LocalComponent, error) {
	_, exists := comps.localComponents[id]
	if exists {
		return nil, fmt.Errorf("component with id '%s' does already exist", id)
	}

	component := newLocalComponent(comps.client, id)
	comps.localComponents[id] = component
	return component, nil
}

func (comps *Components) GetLocalComponent(id string) LocalComponent {
	return comps.localComponents[id]
}

func (comps *Components) RemoveLocalComponent(id string) {
	component := comps.localComponents[id]

	component.Terminate()
	delete(comps.localComponents, id)
}

func (comps *Components) TrackRemoteComponent(remoteInstanceName string, id string) RemoteComponent {
	component := newRemoteComponent(comps.client, remoteInstanceName, id)
	comps.remoteComponents[component] = struct{}{}
	return component
}

func (comps *Components) UntrackRemoteComponent(remoteComponent RemoteComponent) {
	component := remoteComponent.(*remoteComponentImpl)
	component.Terminate()
	delete(comps.remoteComponents, component)
}

type dispatcher struct {
	client        *client
	instanceName  string
	componentId   string
	subscriptions map[string]func([]byte)
	msgToken      tools.RegistrationToken
}

func newDispatcher(client *client, instanceName string, componentId string) *dispatcher {
	disp := &dispatcher{
		client:        client,
		instanceName:  instanceName,
		componentId:   componentId,
		subscriptions: make(map[string]func([]byte)),
	}

	disp.msgToken = disp.client.OnMessage().Register(disp.onMessage)

	return disp
}

func (disp *dispatcher) onMessage(m *message) {
	if m.InstanceName() != disp.instanceName || m.Domain() != componentsDomain {
		return
	}

	parts := strings.Split(m.Path(), "/")
	if len(parts) != 2 {
		return
	}

	componentId := parts[0]
	memberName := parts[1]

	if componentId != disp.componentId {
		return
	}

	handler, exists := disp.subscriptions[memberName]
	if !exists {
		return
	}

	handler(m.Payload())
}

func (disp *dispatcher) AddSubscription(member string, handler func([]byte)) {
	if _, exists := disp.subscriptions[member]; exists {
		panic(fmt.Errorf("member '%s' already registered", member))
	}

	disp.subscriptions[member] = handler

	disp.client.SubscribeNoWait(disp.buildTopic(member))
}

func (disp *dispatcher) Terminate() {
	disp.client.OnMessage().Unregister(disp.msgToken)

	if len(disp.subscriptions) > 0 {
		topics := make([]string, 0)

		for member := range disp.subscriptions {
			topics = append(topics, disp.buildTopic(member))
		}

		disp.client.UnsubscribeNoWait(topics...)
	}
}

func (disp *dispatcher) Emit(memberName string, value []byte, persistent bool) {
	topic := disp.buildTopic(memberName)

	disp.client.PublishNoWait(topic, value, persistent)
}

func (disp *dispatcher) buildTopic(member string) string {
	return disp.client.BuildRemoteTopic(disp.instanceName, componentsDomain, disp.componentId, member)
}

var _ LocalComponent = (*localComponentImpl)(nil)

type localComponentImpl struct {
	disp *dispatcher
}

func newLocalComponent(client *client, id string) *localComponentImpl {
	return &localComponentImpl{
		disp: newDispatcher(client, client.InstanceName(), id),
	}
}

func (comp *localComponentImpl) Terminate() {
	comp.disp.Terminate()
}

func (comp *localComponentImpl) RegisterAction(name string, handler func([]byte)) {
	comp.disp.AddSubscription(name, handler)
}

func (comp *localComponentImpl) SetState(name string, value []byte) {
	comp.disp.Emit(name, value, true)
}

var _ RemoteComponent = (*remoteComponentImpl)(nil)

type remoteComponentImpl struct {
	disp *dispatcher
}

func newRemoteComponent(client *client, remoteInstanceName string, id string) *remoteComponentImpl {
	return &remoteComponentImpl{
		disp: newDispatcher(client, remoteInstanceName, id),
	}
}

func (comp *remoteComponentImpl) Terminate() {
	comp.disp.Terminate()
}

func (comp *remoteComponentImpl) EmitAction(name string, value []byte) {
	comp.disp.Emit(name, value, false)
}

func (comp *remoteComponentImpl) RegisterStateChange(name string, handler func([]byte)) {
	comp.disp.AddSubscription(name, handler)
}
