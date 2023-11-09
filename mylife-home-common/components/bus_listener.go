package components

import (
	"mylife-home-common/bus"
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
	"strings"
	"sync"
)

type BusListener interface {
	Terminate()
}

var _ BusListener = (*busListener)(nil)

type busListener struct {
	transport              *bus.Transport
	registry               Registry
	onChange               chan *bus.InstancePresenceChange
	onChangeListenerClosed chan struct{}
	instances              map[string]*busListenerInstance
	mux                    sync.Mutex
}

// Publish remote components/plugins in the registry
func ListenBus(transport *bus.Transport, registry Registry) BusListener {
	listener := &busListener{
		transport:              transport,
		registry:               registry,
		onChange:               make(chan *bus.InstancePresenceChange),
		onChangeListenerClosed: make(chan struct{}),
		instances:              make(map[string]*busListenerInstance),
	}

	go listener.onChangeListener()
	listener.transport.Presence().OnChange().Subscribe(listener.onChange)

	// Note: a change may occur before we get the complete list
	for _, instanceName := range transport.Presence().GetOnlines() {
		listener.setInstance(instanceName)
	}

	return listener
}

func (listener *busListener) Terminate() {
	listener.transport.Presence().OnChange().Unsubscribe(listener.onChange)
	close(listener.onChange)
	<-listener.onChangeListenerClosed

	// Finish all instances
	// No need to take mutex since the listener routine is terminated
	for instanceName := range listener.instances {
		listener.clearInstance(instanceName)
	}
}

func (listener *busListener) onChangeListener() {
	defer close(listener.onChangeListenerClosed)

	for change := range listener.onChange {
		if change.Online() {
			listener.setInstance(change.InstanceName())
		} else {
			listener.clearInstance(change.InstanceName())
		}
	}
}

func (listener *busListener) setInstance(instanceName string) {
	listener.mux.Lock()
	defer listener.mux.Unlock()

	if _, exists := listener.instances[instanceName]; exists {
		return
	}

	instance := newBusListenerInstance(listener.transport, listener.registry, instanceName)
	listener.instances[instanceName] = instance
}

func (listener *busListener) clearInstance(instanceName string) {
	listener.mux.Lock()
	defer listener.mux.Unlock()

	instance, exists := listener.instances[instanceName]
	if !exists {
		return
	}

	instance.Terminate()
	delete(listener.instances, instanceName)
}

type busListenerInstance struct {
	transport    *bus.Transport
	registry     Registry
	instanceName string
	exit         chan struct{}
	plugins      map[string]struct{}
	components   map[string]struct{}
	// There is no garanty of plugins vs components metadata subscribe results
	// when we connect and the other instance is already here.
	// So in case we get components before plugins, we put them in a pending list
	pendingComponents map[string]*metadata.Component
	mux               sync.Mutex
}

func newBusListenerInstance(transport *bus.Transport, registry Registry, instanceName string) *busListenerInstance {
	instance := &busListenerInstance{
		transport:         transport,
		registry:          registry,
		instanceName:      instanceName,
		exit:              make(chan struct{}),
		plugins:           make(map[string]struct{}),
		components:        make(map[string]struct{}),
		pendingComponents: make(map[string]*metadata.Component),
	}

	go instance.worker()

	return instance
}

// Returns immediately, proper close happens in background
func (instance *busListenerInstance) Terminate() {
	close(instance.exit)

	instance.mux.Lock()
	defer instance.mux.Unlock()

	// immediately clean stuff
	clear(instance.pendingComponents)

	for id := range instance.components {
		instance.clearComponent(id)
	}

	for id := range instance.plugins {
		instance.clearPlugin(id)
	}
}

// Needed to manage lifetime properly (since calls may be long, this outlive instance methods calls)
func (instance *busListenerInstance) worker() {
	view, err := instance.transport.Metadata().CreateView(instance.instanceName)
	if err != nil {
		logger.WithError(err).Errorf("Could not listen for instance metadata '%s'", instance.instanceName)
		return
	}

	defer func() {
		if view != nil {
			instance.transport.Metadata().CloseView(view)
		}
	}()

	onChange := make(chan *bus.ValueChange)
	onChangeListenerClosed := make(chan struct{})

	go instance.onChangeListener(onChange, onChangeListenerClosed)
	view.OnChange().Subscribe(onChange)

	defer func() {
		view.OnChange().Unsubscribe(onChange)
		close(onChange)
		<-onChangeListenerClosed
	}()

	// Note: a change may occur before we get the complete list
	instance.initValues(view.Values())

	// Wait for close
	<-instance.exit
}

func (instance *busListenerInstance) onChangeListener(onChange <-chan *bus.ValueChange, exit chan<- struct{}) {
	defer close(exit)

	for change := range onChange {
		instance.handleChange(change)
	}
}

func (instance *busListenerInstance) handleChange(change *bus.ValueChange) {
	instance.mux.Lock()
	defer instance.mux.Unlock()

	typ, id := instance.parsePath(change.Path())

	switch change.Type() {
	case bus.ValueSet:
		switch typ {
		case "plugins":
			instance.setPlugin(id, change.Value())
		case "components":
			instance.setComponent(id, change.Value())
		}

	case bus.ValueClear:
		switch typ {
		case "plugins":
			instance.clearPlugin(id)
		case "components":
			instance.clearComponent(id)
		}
	}
}

func (instance *busListenerInstance) initValues(values map[string]any) {
	// At least ensure the changeset is consistent
	instance.mux.Lock()
	defer instance.mux.Unlock()

	for path, value := range values {
		typ, id := instance.parsePath(path)

		switch typ {
		case "plugins":
			instance.setPlugin(id, value)
		case "components":
			instance.setComponent(id, value)
		}
	}
}

func (instance *busListenerInstance) parsePath(path string) (typ string, id string) {
	parts := strings.SplitN(path, "/", 2)
	typ = parts[0]
	id = ""
	if len(parts) > 1 {
		id = parts[1]
	}
	return
}

func (instance *busListenerInstance) setPlugin(id string, value any) {
	// set semantic
	if instance.registry.HasPlugin(instance.instanceName, id) {
		return
	}

	plugin := metadata.Serializer.DeserializePlugin(value)

	instance.registry.AddPlugin(instance.instanceName, plugin)
	instance.plugins[id] = struct{}{}

	// See if we have pending component matching this plugin
	for id, netComp := range instance.pendingComponents {
		if netComp.Plugin() == id {
			instance.setComp(netComp.Id(), plugin)
			delete(instance.pendingComponents, id)
		}
	}
}

func (instance *busListenerInstance) clearPlugin(id string) {
	plugin := instance.registry.GetPlugin(instance.instanceName, id)
	// set semantic
	if plugin == nil {
		return
	}

	instance.registry.RemovePlugin(instance.instanceName, plugin)
	delete(instance.plugins, id)
}

func (instance *busListenerInstance) setComponent(id string, value any) {
	// set semantic
	if instance.registry.HasComponent(id) {
		return
	}

	netComp := metadata.Serializer.DeserializeComponent(value)

	plugin := instance.registry.GetPlugin(instance.instanceName, netComp.Plugin())
	if plugin == nil {
		logger.Debugf("Got component '%s' without corresponding plugin '%s' on instance '%s'. Adding to pending list.", netComp.Id(), netComp.Plugin(), instance.instanceName)
		instance.pendingComponents[netComp.Id()] = netComp
		return
	}

	instance.setComp(id, plugin)
}

func (instance *busListenerInstance) setComp(id string, plugin *metadata.Plugin) {
	comp := newBusListenerComponent(instance.transport, instance.instanceName, instance.registry, id, plugin)

	instance.registry.AddComponent(instance.instanceName, comp)
	instance.components[id] = struct{}{}
}

func (instance *busListenerInstance) clearComponent(id string) {
	comp := instance.registry.GetComponent(id)
	// set semantic
	if comp == nil {
		return
	}

	instance.registry.RemoveComponent(instance.instanceName, comp)
	delete(instance.components, id)

	comp.(*busListenerComponent).Terminate()
}

var _ Component = (*busListenerComponent)(nil)

type busListenerComponent struct {
	transport       *bus.Transport
	remoteComponent bus.RemoteComponent
	instanceName    string
	id              string
	plugin          *metadata.Plugin
	state           map[string]tools.ObservableValue[any]
	actions         map[string]chan<- any
}

func newBusListenerComponent(transport *bus.Transport, instanceName string, registry Registry, id string, plugin *metadata.Plugin) *busListenerComponent {
	comp := &busListenerComponent{
		transport:    transport,
		instanceName: instanceName,
		id:           id,
		plugin:       plugin,
		state:        make(map[string]tools.ObservableValue[any]),
		actions:      make(map[string]chan<- any),
	}

	comp.remoteComponent = transport.Components().TrackRemoteComponent(comp.instanceName, comp.id)

	for _, name := range comp.plugin.MemberNames() {
		member := comp.plugin.Member(name)
		switch member.MemberType() {
		case metadata.State:
			comp.state[name] = comp.initState(member)

		case metadata.Action:
			comp.actions[name] = comp.initAction(member)
		}
	}

	return comp
}

func (comp *busListenerComponent) initState(member *metadata.Member) tools.ObservableValue[any] {
	name := member.Name()
	// Note: nil = no value for now, will come as soon as we get them from the bus
	subject := tools.MakeSubjectValue[any](nil)

	// finish init in background
	go comp.remoteComponent.RegisterStateChange(name, func(data []byte) {
		value := bus.Encoding.ReadValue(member.ValueType(), data)
		// dispatch async
		go subject.Update(value)
	})

	return subject
}

func (comp *busListenerComponent) initAction(member *metadata.Member) chan<- any {
	name := member.Name()
	channel := make(chan any)

	tools.DispatchChannel(channel, func(value any) {
		data := bus.Encoding.WriteValue(member.ValueType(), value)
		// emit async
		go comp.remoteComponent.EmitAction(name, data)
	})

	return channel
}

func (comp *busListenerComponent) Terminate() {

	for _, channel := range comp.actions {
		close(channel)
	}

	// finish close in background
	go comp.transport.Components().UntrackRemoteComponent(comp.remoteComponent)
}

func (comp *busListenerComponent) Id() string {
	return comp.id
}

func (comp *busListenerComponent) Plugin() *metadata.Plugin {
	return comp.plugin
}

func (comp *busListenerComponent) StateItem(name string) tools.ObservableValue[any] {
	return comp.state[name]
}

func (comp *busListenerComponent) Action(name string) chan<- any {
	return comp.actions[name]
}
