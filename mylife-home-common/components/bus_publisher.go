package components

import (
	"mylife-home-common/bus"
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"

	"golang.org/x/exp/maps"
)

const registryLocalInstanceName = ""

type BusPublisher interface {
	Terminate()
}

var _ BusPublisher = (*busPublisher)(nil)

type busPublisher struct {
	transport              *bus.Transport
	registry               Registry
	busPublisherPlugins    map[*metadata.Plugin]*busPublisherPlugin
	busPublisherComponents map[Component]*busPublisherComponent
	pluginChangeChan       chan *PluginChange
	componentChangeChan    chan *ComponentChange
	onlineChan             chan bool
}

// Publish registry local components/plugins to the bus
func PublishBus(transport *bus.Transport, registry Registry) BusPublisher {
	publisher := &busPublisher{
		transport:              transport,
		registry:               registry,
		busPublisherPlugins:    make(map[*metadata.Plugin]*busPublisherPlugin),
		busPublisherComponents: make(map[Component]*busPublisherComponent),
		pluginChangeChan:       make(chan *PluginChange),
		componentChangeChan:    make(chan *ComponentChange),
		onlineChan:             make(chan bool),
	}

	go publisher.worker()

	publisher.transport.Online().Subscribe(publisher.onlineChan, false)
	publisher.registry.OnPluginChange().Subscribe(publisher.pluginChangeChan)
	publisher.registry.OnComponentChange().Subscribe(publisher.componentChangeChan)

	return publisher
}

func (publisher *busPublisher) Terminate() {
	publisher.transport.Online().Unsubscribe(publisher.onlineChan)
	publisher.registry.OnPluginChange().Unsubscribe(publisher.pluginChangeChan)
	publisher.registry.OnComponentChange().Unsubscribe(publisher.componentChangeChan)

	close(publisher.onlineChan)
	close(publisher.pluginChangeChan)
	close(publisher.componentChangeChan)
}

func (publisher *busPublisher) worker() {
	publisher.onInit()
	defer publisher.onClose()

	for {
		select {
		case online, ok := <-publisher.onlineChan:
			if !ok {
				return // closing
			}

			publisher.onOnlineChange(online)

		case change, ok := <-publisher.pluginChangeChan:
			if !ok {
				return // closing
			}

			publisher.onPluginChange(change)

		case change, ok := <-publisher.componentChangeChan:
			if !ok {
				return // closing
			}

			publisher.onComponentChange(change)
		}
	}
}

func (publisher *busPublisher) onInit() {
	for _, plugin := range publisher.registry.GetPlugins(registryLocalInstanceName) {
		publisher.publishPlugin(plugin)
	}

	for _, componentData := range publisher.registry.GetComponentsData() {
		if componentData.InstanceName() != registryLocalInstanceName {
			continue
		}

		publisher.publishComponent(componentData.Component())
	}
}

func (publisher *busPublisher) onClose() {
	for _, component := range maps.Keys(publisher.busPublisherComponents) {
		publisher.unpublishComponent(component)
	}

	for _, plugin := range maps.Keys(publisher.busPublisherPlugins) {
		publisher.unpublishPlugin(plugin)
	}
}

func (publisher *busPublisher) onPluginChange(change *PluginChange) {
	if change.InstanceName() != registryLocalInstanceName {
		return
	}

	switch change.Action() {
	case RegistryAdd:
		publisher.publishPlugin(change.Plugin())
	case RegistryRemove:
		publisher.unpublishPlugin(change.Plugin())
	}
}

func (publisher *busPublisher) onComponentChange(change *ComponentChange) {
	if change.InstanceName() != registryLocalInstanceName {
		return
	}

	switch change.Action() {
	case RegistryAdd:
		publisher.publishComponent(change.Component())
	case RegistryRemove:
		publisher.unpublishComponent(change.Component())
	}
}

func (publisher *busPublisher) publishPlugin(plugin *metadata.Plugin) {
	bp := newBusPublisherPlugin(plugin, &busPublisherApi{publisher.transport})
	publisher.busPublisherPlugins[plugin] = bp
}

func (publisher *busPublisher) unpublishPlugin(plugin *metadata.Plugin) {
	bp := publisher.busPublisherPlugins[plugin]
	bp.close()
	delete(publisher.busPublisherPlugins, plugin)
}

func (publisher *busPublisher) publishComponent(component Component) {
	bc := newBusPublisherComponent(component, &busPublisherApi{publisher.transport})
	publisher.busPublisherComponents[component] = bc
}

func (publisher *busPublisher) unpublishComponent(component Component) {
	bc := publisher.busPublisherComponents[component]
	bc.close()
	delete(publisher.busPublisherComponents, component)
}

func (publisher *busPublisher) onOnlineChange(online bool) {
	plugins := maps.Values(publisher.busPublisherPlugins)
	components := maps.Values(publisher.busPublisherComponents)

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

type busPublisherApi struct {
	transport *bus.Transport
}

func (api *busPublisherApi) PublishMeta(path string, value any) {
	go func() {
		if err := api.transport.Metadata().Set(path, value); err != nil {
			logger.WithError(err).Errorf("Could not publish metadata '%s'", path)
		}
	}()
}

func (api *busPublisherApi) UnpublishMeta(path string) {
	go func() {
		if err := api.transport.Metadata().Clear(path); err != nil {
			logger.WithError(err).Errorf("Could not unpublish metadata '%s'", path)
		}
	}()
}

func (api *busPublisherApi) IsOnline() bool {
	return api.transport.Online().Get()
}

func (api *busPublisherApi) AddLocalComponent(id string) bus.LocalComponent {
	return api.transport.Components().AddLocalComponent(id)
}

func (api *busPublisherApi) RemoveLocalComponent(comp bus.LocalComponent) {
	go func() {
		api.transport.Components().RemoveLocalComponent(comp)
	}()
}

type busPublisherPlugin struct {
	path string
	meta any
	api  *busPublisherApi
}

func newBusPublisherPlugin(plugin *metadata.Plugin, api *busPublisherApi) *busPublisherPlugin {
	// metadata does not change, we can build it directly
	bp := &busPublisherPlugin{
		path: "plugins/" + plugin.Id(),
		meta: metadata.Serializer.SerializePlugin(plugin),
		api:  api,
	}

	if bp.api.IsOnline() {
		bp.publishMeta()
	}

	return bp
}

func (bp *busPublisherPlugin) onOnlineChange(online bool) {
	if online {
		bp.publishMeta()
	}
}

func (bp *busPublisherPlugin) close() {
	if bp.api.IsOnline() {
		bp.unpublishMeta()
	}
}

func (bp *busPublisherPlugin) publishMeta() {
	bp.api.PublishMeta(bp.path, bp.meta)
}

func (bp *busPublisherPlugin) unpublishMeta() {
	bp.api.UnpublishMeta(bp.path)
}

type busPublisherComponent struct {
	path               string
	meta               any
	api                *busPublisherApi
	component          Component
	transportComponent bus.LocalComponent
	stateChans         map[string]chan any
}

func newBusPublisherComponent(component Component, api *busPublisherApi) *busPublisherComponent {
	// metadata does not change, we can build it directly
	metaComp := metadata.MakeComponent(component.Id(), component.Plugin().Id())
	bc := &busPublisherComponent{
		path:       "components/" + component.Id(),
		meta:       metadata.Serializer.SerializeComponent(metaComp),
		api:        api,
		component:  component,
		stateChans: make(map[string]chan any),
	}

	if bc.api.IsOnline() {
		bc.publishComponent()
		bc.publishMeta()
	}

	return bc
}

func (bc *busPublisherComponent) onOnlineChange(online bool) {
	if online {
		bc.publishComponent()
		bc.publishMeta()
	} else {
		bc.unpublishComponent()
	}
}

func (bc *busPublisherComponent) close() {
	if bc.api.IsOnline() {
		bc.unpublishMeta()
		bc.unpublishComponent()
	}
}

func (bc *busPublisherComponent) publishState(name string, value any) {
	member := bc.component.Plugin().Member(name)
	data := bus.Encoding.WriteValue(member.ValueType(), value)

	go bc.transportComponent.SetState(name, data)
}

func (bc *busPublisherComponent) publishMeta() {
	bc.api.PublishMeta(bc.path, bc.meta)
}

func (bc *busPublisherComponent) unpublishMeta() {
	bc.api.UnpublishMeta(bc.path)
}

func (bc *busPublisherComponent) publishComponent() {
	bc.transportComponent = bc.api.AddLocalComponent(bc.component.Id())

	plugin := bc.component.Plugin()
	for _, name := range plugin.MemberNames() {
		member := plugin.Member(name)

		switch member.MemberType() {
		case metadata.Action:
			bc.registerAction(member)

		case metadata.State:
			bc.registerState(member)
		}
	}
}

func (bc *busPublisherComponent) registerAction(member *metadata.Member) {
	name := member.Name()
	typ := member.ValueType()
	channel := bc.component.Action(name)

	// register is blocking
	go bc.transportComponent.RegisterAction(name, func(data []byte) {
		value := bus.Encoding.ReadValue(typ, data)
		go func() {
			channel <- value
		}()
	})
}

func (bc *busPublisherComponent) registerState(member *metadata.Member) {
	name := member.Name()
	channel := make(chan any)
	bc.stateChans[name] = channel

	tools.DispatchChannel(channel, func(value any) {
		bc.publishState(name, value)
	})

	obsvalue := bc.component.StateItem(name)
	obsvalue.Subscribe(channel, true)
}

func (bc *busPublisherComponent) unpublishComponent() {
	bc.api.RemoveLocalComponent(bc.transportComponent)

	for name, channel := range bc.stateChans {
		obsvalue := bc.component.StateItem(name)
		obsvalue.Unsubscribe(channel)

		close(channel)

		delete(bc.stateChans, name)
	}
}
