package bus

const componentsDomain = "components"

type LocalComponent interface {
	// Note: blocking
	RegisterAction(name string, handler func([]byte))
	// Note: blocking
	SetState(name string, value []byte)
}

type RemoteComponent interface {
	// Note: blocking
	EmitAction(name string, value []byte)
	// Note: blocking
	RegisterStateChange(name string, handler func([]byte))
}

type Components struct {
	client *client
}

func newComponents(client *client) *Components {
	return &Components{
		client: client,
	}
}

// Note: automatically dropped on disconnection
func (comps *Components) AddLocalComponent(id string) LocalComponent {
	return newLocalComponent(comps.client, id)
}

// Note: blocking if connected
func (comps *Components) RemoveLocalComponent(comp LocalComponent) {
	component := comp.(*localComponent)

	component.Terminate()
}

// Note: automatically dropped on disconnection
func (comps *Components) TrackRemoteComponent(remoteInstanceName string, id string) RemoteComponent {
	return newRemoteComponent(comps.client, remoteInstanceName, id)
}

// Note: blocking if connected
func (comps *Components) UntrackRemoteComponent(remoteComp RemoteComponent) {
	component := remoteComp.(*remoteComponent)
	component.Terminate()
}

type dispatcher struct {
	client       *client
	instanceName string
	componentId  string
	topics       []string
}

func newDispatcher(client *client, instanceName string, componentId string) *dispatcher {
	return &dispatcher{
		client:       client,
		instanceName: instanceName,
		componentId:  componentId,
		topics:       make([]string, 0),
	}
}

// Note: blocking
func (disp *dispatcher) AddSubscription(member string, handler func([]byte)) {
	topic := disp.buildTopic(member)

	cb := func(m *message) {
		handler(m.Payload())
	}

	if err := disp.client.Subscribe(topic, cb); err != nil {
		logger.WithError(err).Errorf("Could not subscribe topics: %+v", disp.topics)
		return
	}

	disp.topics = append(disp.topics, topic)
}

// Note: blocking
func (disp *dispatcher) Terminate() {
	if len(disp.topics) > 0 {
		if err := disp.client.Unsubscribe(disp.topics...); err != nil {
			logger.WithError(err).Errorf("Could not unsubscribe topics: %+v", disp.topics)
		}
	}
}

// Note: blocking
func (disp *dispatcher) Emit(memberName string, value []byte, persistent bool) {
	topic := disp.buildTopic(memberName)

	var err error
	if persistent {
		err = disp.client.PublishRetain(topic, value)
	} else {
		err = disp.client.Publish(topic, value)
	}

	if err != nil {
		logger.WithError(err).Errorf("Could not publish topic: %s", topic)
	}
}

func (disp *dispatcher) buildTopic(member string) string {
	return disp.client.BuildRemoteTopic(disp.instanceName, componentsDomain, disp.componentId, member)
}

var _ LocalComponent = (*localComponent)(nil)

type localComponent struct {
	disp *dispatcher
}

func newLocalComponent(client *client, id string) *localComponent {
	return &localComponent{
		disp: newDispatcher(client, client.InstanceName(), id),
	}
}

func (comp *localComponent) Terminate() {
	comp.disp.Terminate()
}

func (comp *localComponent) RegisterAction(name string, handler func([]byte)) {
	comp.disp.AddSubscription(name, handler)
}

func (comp *localComponent) SetState(name string, value []byte) {
	comp.disp.Emit(name, value, true)
}

var _ RemoteComponent = (*remoteComponent)(nil)

type remoteComponent struct {
	disp *dispatcher
}

func newRemoteComponent(client *client, remoteInstanceName string, id string) *remoteComponent {
	return &remoteComponent{
		disp: newDispatcher(client, remoteInstanceName, id),
	}
}

func (comp *remoteComponent) Terminate() {
	comp.disp.Terminate()
}

func (comp *remoteComponent) EmitAction(name string, value []byte) {
	comp.disp.Emit(name, value, false)
}

func (comp *remoteComponent) RegisterStateChange(name string, handler func([]byte)) {
	comp.disp.AddSubscription(name, handler)
}
