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
type busPublisher struct {
	transport   *bus.Transport
	registry    *Registry
	instances   map[string]*busInstance
	changeToken tools.RegistrationToken
}

func newBusPublisher(transport *bus.Transport, registry *Registry) *busPublisher {
	if !transport.Presence().Tracking() {
		panic("cannot use 'BusPublisher' with presence tracking disabled")
	}

	publisher := &busPublisher{
		transport: transport,
		registry:  registry,
		instances: make(map[string]*busInstance), // only changed from mqtt thread
	}

	publisher.changeToken = publisher.transport.Presence().OnInstanceChange().Register(publisher.onInstanceChange)

	for _, instanceName := range transport.Presence().GetOnlines() {
		publisher.setInstance(instanceName)
	}

	return publisher
}

func (publisher *busPublisher) Terminate() {
	publisher.transport.Presence().OnInstanceChange().Unregister(publisher.changeToken)

	// clone for stability
	for _, instanceName := range maps.Keys(publisher.instances) {
		publisher.clearInstance(instanceName)
	}
}

func (publisher *busPublisher) onInstanceChange(change *bus.InstancePresenceChange) {
	if change.Online() {
		publisher.setInstance(change.InstanceName())
	} else {
		publisher.clearInstance(change.InstanceName())
	}
}

func (publisher *busPublisher) setInstance(instanceName string) {
	instance := newBusInstance(publisher.transport, publisher.registry, instanceName)
	publisher.instances[instanceName] = instance
}

func (publisher *busPublisher) clearInstance(instanceName string) {
	instance := publisher.instances[instanceName]
	instance.Terminate()
	delete(publisher.instances, instanceName)
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

	fireAndForget(func() error {
		view, err := instance.transport.Metadata().CreateView(instance.instanceName)
		if err != nil {
			return err
		}

		instance.view = view
		instance.viewChangeToken = instance.view.OnChange().Register(instance.onViewChange)

		instance.initView()

		return nil
	})

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

	fireAndForget(func() error {
		instance.view.OnChange().Unregister(instance.viewChangeToken)

		instance.transport.Metadata().CloseView(instance.view)

		return nil
	})
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

	plugin := metadata.Seralizer.DeserializePlugin(value)

	instance.registry.AddPlugin(instance.instanceName, plugin)
}

func (instance *busInstance) setComponent(id string, value any) {
	// set semantic
	if instance.registry.HasComponent(id) {
		return
	}

	netComp := metadata.Seralizer.DeserializeComponent(value)
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

	fireAndForget(func() error {
		for name := range comp.state {
			// Else it seems that the name inside the closure in incorrect
			closedName := name
			err := comp.remoteComponent.RegisterStateChange(name, func(data []byte) {
				comp.stateChange(closedName, data)
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	return comp
}

func (comp *busComponent) Terminate() {
	fireAndForget(func() error {
		comp.transport.Components().UntrackRemoteComponent(comp.remoteComponent)
		return nil
	})
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
	fireAndForget(func() error {
		return comp.remoteComponent.EmitAction(name, data)
	})
}

func fireAndForget(callback func() error) {
	go func() {
		if err := callback(); err != nil {
			logger.WithError(err).Error("Fire and forget failed")
		}
	}()
}
