package manager

import (
	"mylife-home-common/bus"
	"mylife-home-common/components"
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
	"sync"

	"golang.org/x/exp/maps"
)

const registryLocalInstanceName = ""

// Publish registry local components/plugins to the bus
type busPublisher struct {
	transport            *bus.Transport
	registry             *components.Registry
	busPlugins           map[*metadata.Plugin]*busPlugin
	busComponents        map[components.Component]*busComponent
	mux                  sync.Mutex
	componentChangeToken tools.RegistrationToken
	pluginChangeToken    tools.RegistrationToken
	onlineChangeToken    tools.RegistrationToken
}

func newBusPublisher(transport *bus.Transport, registry *components.Registry) *busPublisher {
	if !transport.Presence().Tracking() {
		panic("cannot use 'BusPublisher' with presence tracking disabled")
	}

	publisher := &busPublisher{
		transport:     transport,
		registry:      registry,
		busPlugins:    make(map[*metadata.Plugin]*busPlugin),
		busComponents: make(map[components.Component]*busComponent),
	}

	publisher.pluginChangeToken = publisher.registry.OnPluginChange().Register(publisher.onPluginChange)
	publisher.componentChangeToken = publisher.registry.OnComponentChange().Register(publisher.onComponentChange)
	publisher.onlineChangeToken = publisher.transport.OnOnlineChanged().Register(publisher.onOnlineChange)

	// clone for stability
	for _, plugin := range publisher.registry.GetPlugins(registryLocalInstanceName).Clone() {
		publisher.publishPlugin(plugin)
	}

	// clone for stability
	for _, componentData := range publisher.registry.GetComponentsData().Clone() {
		if componentData.InstanceName() != registryLocalInstanceName {
			continue
		}

		publisher.publishComponent(componentData.Component())
	}

	return publisher
}

func (publisher *busPublisher) Terminate() {
	publisher.registry.OnPluginChange().Unregister(publisher.pluginChangeToken)
	publisher.registry.OnComponentChange().Unregister(publisher.componentChangeToken)
	publisher.transport.OnOnlineChanged().Unregister(publisher.pluginChangeToken)

	// no need for lock anymore since there is no event left
	for _, component := range maps.Keys(publisher.busComponents) {
		publisher.unpublishComponent(component)
	}

	for _, plugin := range maps.Keys(publisher.busPlugins) {
		publisher.unpublishPlugin(plugin)
	}
}

func (publisher *busPublisher) onPluginChange(change *components.PluginChange) {
	if change.InstanceName() != registryLocalInstanceName {
		return
	}

	switch change.Action() {
	case components.RegistryAdd:
		publisher.publishPlugin(change.Plugin())
	case components.RegistryRemove:
		publisher.unpublishPlugin(change.Plugin())
	}
}

func (publisher *busPublisher) onComponentChange(change *components.ComponentChange) {
	if change.InstanceName() != registryLocalInstanceName {
		return
	}

	switch change.Action() {
	case components.RegistryAdd:
		publisher.publishComponent(change.Component())
	case components.RegistryRemove:
		publisher.unpublishComponent(change.Component())
	}
}

func (publisher *busPublisher) publishPlugin(plugin *metadata.Plugin) {
	publisher.mux.Lock()
	defer publisher.mux.Unlock()

	bp := newBusPlugin(plugin, publisher.transport)
	publisher.busPlugins[plugin] = bp
}

func (publisher *busPublisher) unpublishPlugin(plugin *metadata.Plugin) {
	publisher.mux.Lock()
	defer publisher.mux.Unlock()

	bp := publisher.busPlugins[plugin]
	bp.close()
	delete(publisher.busPlugins, plugin)
}

func (publisher *busPublisher) publishComponent(component components.Component) {
	publisher.mux.Lock()
	defer publisher.mux.Unlock()

	bc := newBusComponent(component, publisher.transport)
	publisher.busComponents[component] = bc
}

func (publisher *busPublisher) unpublishComponent(component components.Component) {
	publisher.mux.Lock()
	defer publisher.mux.Unlock()

	bc := publisher.busComponents[component]
	bc.close()
	delete(publisher.busComponents, component)
}

func (publisher *busPublisher) onOnlineChange(online bool) {
	publisher.mux.Lock()
	plugins := maps.Values(publisher.busPlugins)
	components := maps.Values(publisher.busComponents)
	publisher.mux.Unlock()

	// process plugins first, then components
	//  - on online : when a component is published, its plugin must already exist
	//  - on offline : we are offline so we don't care, let's do it in the same order
	for _, bp := range plugins {
		bp.onOnlineChange(online)
	}

	for _, bc := range components {
		bc.onOnlineChange(online)
	}
}

type busPlugin struct {
	path      string
	meta      any
	transport *bus.Transport
}

func newBusPlugin(plugin *metadata.Plugin, transport *bus.Transport) *busPlugin {
	// metadata does not change, we can build it directly
	bp := &busPlugin{
		path:      "plugins/" + plugin.Id(),
		meta:      metadata.Serializer.SerializePlugin(plugin),
		transport: transport,
	}

	if bp.transport.Online() {
		bp.publishMeta()
	}

	return bp
}

func (bp *busPlugin) onOnlineChange(online bool) {
	if online {
		bp.publishMeta()
	}
}

func (bp *busPlugin) close() {
	if bp.transport.Online() {
		bp.unpublishMeta()
	}
}

func (bp *busPlugin) publishMeta() {
	bp.transport.Metadata().Set(bp.path, bp.meta)
}

func (bp *busPlugin) unpublishMeta() {
	bp.transport.Metadata().Clear(bp.path)
}

type busComponent struct {
	path               string
	meta               any
	transport          *bus.Transport
	component          components.Component
	changeToken        tools.RegistrationToken
	transportComponent bus.LocalComponent
}

func newBusComponent(component components.Component, transport *bus.Transport) *busComponent {
	// metadata does not change, we can build it directly
	metaComp := metadata.MakeComponent(component.Id(), component.Plugin().Id())
	bc := &busComponent{
		path:      "components/" + component.Id(),
		meta:      metadata.Serializer.SerializeComponent(metaComp),
		transport: transport,
		component: component,
	}

	bc.changeToken = bc.component.OnStateChange().Register(bc.onStateChange)

	if bc.transport.Online() {
		bc.publishMeta()
		bc.publishComponent()
	}

	return bc
}

func (bc *busComponent) onOnlineChange(online bool) {
	if online {
		bc.publishComponent()
		bc.publishMeta()
	} else {
		bc.unpublishComponent()
	}
}

func (bc *busComponent) close() {
	bc.component.OnStateChange().Unregister(bc.changeToken)

	if bc.transport.Online() {
		bc.unpublishMeta()
		bc.unpublishComponent()
	}
}

func (bc *busComponent) onStateChange(change *components.StateChange) {
	if !bc.transport.Online() {
		return
	}

	name := change.Name()
	value := change.Value()

	bc.publishState(name, value)
}

func (bc *busComponent) publishState(name string, value any) {
	member := bc.component.Plugin().Member(name)
	data := bus.Encoding.WriteValue(member.ValueType(), value)

	fireAndForget(func() error {
		return bc.transportComponent.SetState(name, data)
	})
}

func (bc *busComponent) publishMeta() {
	bc.transport.Metadata().Set(bc.path, bc.meta)
}

func (bc *busComponent) unpublishMeta() {
	bc.transport.Metadata().Clear(bc.path)
}

func (bc *busComponent) publishComponent() {
	transportComponent, err := bc.transport.Components().AddLocalComponent(bc.component.Id())
	if err != nil {
		logger.Errorf("Could not publish local component '%s' on bus: %s", bc.component.Id(), err)
		return
	}

	bc.transportComponent = transportComponent

	plugin := bc.component.Plugin()
	for _, name := range plugin.MemberNames() {
		member := plugin.Member(name)

		switch member.MemberType() {
		case metadata.Action:
			bc.registerAction(name, member)

		case metadata.State:
			value := bc.component.GetStateItem(name)
			bc.publishState(name, value)
		}
	}
}

func (bc *busComponent) registerAction(name string, member *metadata.Member) {
	typ := member.ValueType()

	fireAndForget(func() error {
		return bc.transportComponent.RegisterAction(name, func(data []byte) {
			value := bus.Encoding.ReadValue(typ, data)
			bc.component.ExecuteAction(name, value)
		})
	})
}

func (bc *busComponent) unpublishComponent() {
	bc.transport.Components().RemoveLocalComponent(bc.component.Id())
}

func fireAndForget(callback func() error) {
	go func() {
		if err := callback(); err != nil {
			logger.WithError(err).Error("Fire and forget failed")
		}
	}()
}
