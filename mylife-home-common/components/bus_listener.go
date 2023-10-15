package components

import (
	"fmt"
	"mylife-home-common/bus"
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
	"strings"
	"sync"

	"golang.org/x/exp/maps"
)

// Publish remote components/plugins in the registry
type busListener struct {
	transport   *bus.Transport
	registry    *Registry
	instances   map[string]*busInstance
	changeToken tools.RegistrationToken
}

func newBusPublisher(transport *bus.Transport, registry *Registry) *busListener {
	if !transport.Presence().Tracking() {
		panic("cannot use 'BusPublisher' with presence tracking disabled")
	}

	listener := &busListener{
		transport: transport,
		registry:  registry,
		instances: make(map[string]*busInstance), // only changed from mqtt thread
	}

	listener.changeToken = listener.transport.Presence().OnInstanceChange().Register(listener.onInstanceChange)

	for _, instanceName := range transport.Presence().GetOnlines() {
		listener.setInstance(instanceName)
	}

	return listener
}

func (listener *busListener) Terminate() {
	listener.transport.Presence().OnInstanceChange().Unregister(listener.changeToken)

	// clone for stability
	for _, instanceName := range maps.Keys(listener.instances) {
		listener.clearInstance(instanceName)
	}
}

func (listener *busListener) onInstanceChange(change *bus.InstancePresenceChange) {
	if change.Online() {
		listener.setInstance(change.InstanceName())
	} else {
		listener.clearInstance(change.InstanceName())
	}
}

func (listener *busListener) setInstance(instanceName string) {
	instance := newBusInstance(listener.transport, listener.registry, instanceName)
	listener.instances[instanceName] = instance
}

func (listener *busListener) clearInstance(instanceName string) {
	instance := listener.instances[instanceName]
	instance.Terminate()
	delete(listener.instances, instanceName)
}

type busInstance struct {
	transport       *bus.Transport
	view            bus.RemoteMetadataView
	viewChangeToken tools.RegistrationToken
	registry        *Registry
	instanceName    string
}

func newBusInstance(transport *bus.Transport, registry *Registry, instanceName string) *busInstance {
	instance := &busInstance{
		transport:    transport,
		registry:     registry,
		instanceName: instanceName,
	}

	view := instance.transport.Metadata().CreateView(instance.instanceName)
	instance.view = view
	instance.viewChangeToken = instance.view.OnChange().Register(instance.onViewChange)
	instance.initView()

	return instance
}

func (instance *busInstance) Terminate() {
	// clone for stability
	for _, data := range instance.registry.GetComponentsData().Clone() {
		if data.InstanceName() == instance.instanceName {
			instance.clearComponent(data.Component().Id())
		}
	}

	for _, plugin := range instance.registry.GetPlugins(instance.instanceName).Clone() {
		instance.clearPlugin(plugin.Id())
	}

	instance.view.OnChange().Unregister(instance.viewChangeToken)
	instance.transport.Metadata().CloseView(instance.view)
}

func (instance *busInstance) initView() {
	// set first plugins then components
	type item struct {
		id    string
		value any
	}

	plugins := make([]item, 0)
	components := make([]item, 0)

	it := instance.view.Values().Iterate()
	for it.Next() {
		path, value := it.Get()

		parts := strings.SplitN(path, "/", 2)
		typ := parts[0]
		id := ""
		if len(parts) > 1 {
			id = parts[1]
		}

		switch typ {
		case "plugins":
			plugins = append(plugins, item{id, value})
		case "components":
			components = append(components, item{id, value})
		}
	}

	for _, it := range plugins {
		instance.setPlugin(it.id, it.value)
	}

	for _, it := range components {
		instance.setComponent(it.id, it.value)
	}
}

func (instance *busInstance) onViewChange(change *bus.ValueChange) {
	parts := strings.SplitN(change.Path(), "/", 2)
	typ := parts[0]
	id := ""
	if len(parts) > 1 {
		id = parts[1]
	}

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

func (instance *busInstance) setPlugin(id string, value any) {
	// set semantic
	if instance.registry.HasPlugin(instance.instanceName, id) {
		return
	}

	plugin := metadata.Serializer.DeserializePlugin(value)

	instance.registry.AddPlugin(instance.instanceName, plugin)
}

func (instance *busInstance) setComponent(id string, value any) {
	// set semantic
	if instance.registry.HasComponent(id) {
		return
	}

	netComp := metadata.Serializer.DeserializeComponent(value)
	comp := newBusComponent(instance.transport, instance.instanceName, instance.registry, netComp)
	instance.registry.AddComponent(instance.instanceName, comp)
}

func (instance *busInstance) clearPlugin(id string) {
	plugin := instance.registry.GetPlugin(instance.instanceName, id)
	instance.registry.RemovePlugin(instance.instanceName, plugin)
}

func (instance *busInstance) clearComponent(id string) {
	comp := instance.registry.GetComponent(id).(*busComponent)
	instance.registry.RemoveComponent(instance.instanceName, comp)
	comp.Terminate()
}

var _ Component = (*busComponent)(nil)

type busComponent struct {
	transport       *bus.Transport
	remoteComponent bus.RemoteComponent
	instanceName    string
	id              string
	plugin          *metadata.Plugin
	state           map[string]any
	stateLock       sync.RWMutex
	onStateChange   *tools.CallbackManager[*StateChange]
}

func newBusComponent(transport *bus.Transport, instanceName string, registry *Registry, netComp *metadata.Component) *busComponent {
	comp := &busComponent{
		transport:     transport,
		instanceName:  instanceName,
		id:            netComp.Id(),
		plugin:        registry.GetPlugin(instanceName, netComp.Plugin()),
		state:         make(map[string]any),
		onStateChange: tools.NewCallbackManager[*StateChange](),
	}

	comp.remoteComponent = transport.Components().TrackRemoteComponent(comp.instanceName, comp.id)

	for _, name := range comp.plugin.MemberNames() {
		member := comp.plugin.Member(name)
		if member.MemberType() == metadata.State {
			comp.state[name] = nil
		}
	}

	for name := range comp.state {
		// Else it seems that the name inside the closure in incorrect
		closedName := name
		comp.remoteComponent.RegisterStateChange(name, func(data []byte) {
			comp.stateChange(closedName, data)
		})
	}

	return comp
}

func (comp *busComponent) Terminate() {
	comp.transport.Components().UntrackRemoteComponent(comp.remoteComponent)
}

func (comp *busComponent) OnStateChange() tools.CallbackRegistration[*StateChange] {
	return comp.onStateChange
}

func (comp *busComponent) Id() string {
	return comp.id
}

func (comp *busComponent) Plugin() *metadata.Plugin {
	return comp.plugin
}

func (comp *busComponent) GetStateItem(name string) any {
	comp.stateLock.RLock()
	defer comp.stateLock.RUnlock()

	return comp.state[name]
}

func (comp *busComponent) GetState() tools.ReadonlyMap[string, any] {
	comp.stateLock.RLock()
	defer comp.stateLock.RUnlock()

	// need to provide a copy to keep stable
	return tools.NewReadonlyMap[string, any](maps.Clone(comp.state))
}

func (comp *busComponent) stateChange(name string, data []byte) {
	comp.stateLock.Lock()
	defer comp.stateLock.Unlock()

	member := comp.plugin.Member(name)
	value := bus.Encoding.ReadValue(member.ValueType(), data)
	comp.state[name] = value
	comp.onStateChange.Execute(&StateChange{
		name:  name,
		value: value,
	})
}

func (comp *busComponent) ExecuteAction(name string, value any) {
	member := comp.plugin.Member(name)
	if member == nil || member.MemberType() != metadata.Action {
		panic(fmt.Errorf("unknown action '%s' on component '%s' (plugin=%s:%s)", name, comp.id, comp.instanceName, comp.plugin.Id()))
	}

	data := bus.Encoding.WriteValue(member.ValueType(), value)
	comp.remoteComponent.EmitAction(name, data)
}
