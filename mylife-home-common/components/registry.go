package components

import (
	"fmt"
	"mylife-home-common/bus"
	"mylife-home-common/components/metadata"
	"mylife-home-common/log"
	"mylife-home-common/tools"
	"sync"
)

var logger = log.CreateLogger("mylife:home:components:registry")

type RegistryAction int

const (
	RegistryAdd    RegistryAction = iota - 1
	RegistryRemove RegistryAction = iota + 1
)

type ComponentChange struct {
	action       RegistryAction
	instanceName string
	component    Component
}

func (change *ComponentChange) Action() RegistryAction {
	return change.action
}

func (change *ComponentChange) InstanceName() string {
	return change.instanceName
}

func (change *ComponentChange) Component() Component {
	return change.component
}

type PluginChange struct {
	action       RegistryAction
	instanceName string
	plugin       *metadata.Plugin
}

func (change *PluginChange) Action() RegistryAction {
	return change.action
}

func (change *PluginChange) InstanceName() string {
	return change.instanceName
}

func (change *PluginChange) Plugin() *metadata.Plugin {
	return change.plugin
}

type RegistryOptions struct {
	publishRemoteComponents bool
	transport               *bus.Transport
}

func NewRegistryOptions() *RegistryOptions {
	return &RegistryOptions{}
}

func (options *RegistryOptions) PublishRemoteComponents(transport *bus.Transport) *RegistryOptions {
	options.publishRemoteComponents = true
	options.transport = transport
	return options
}

type instanceData struct {
	plugins    map[*metadata.Plugin]struct{}
	components map[Component]struct{}
}

type ComponentData struct {
	instanceName string
	component    Component
}

func (data *ComponentData) InstanceName() string {
	return data.instanceName
}

func (data *ComponentData) Component() Component {
	return data.component
}

type Registry struct {
	onComponentChange  *tools.CallbackManager[*ComponentChange]
	onPluginChange     *tools.CallbackManager[*PluginChange]
	lock               sync.RWMutex
	components         map[string]*ComponentData
	pluginsPerInstance map[string]*metadata.Plugin
	instances          map[string]*instanceData
	publisher          *busPublisher
}

func NewRegistry(options *RegistryOptions) *Registry {
	registry := &Registry{
		onComponentChange:  tools.NewCallbackManager[*ComponentChange](),
		onPluginChange:     tools.NewCallbackManager[*PluginChange](),
		components:         make(map[string]*ComponentData),
		pluginsPerInstance: make(map[string]*metadata.Plugin),
		instances:          make(map[string]*instanceData),
	}

	if options.publishRemoteComponents {
		registry.publisher = newBusPublisher(options.transport, registry)
	}

	return registry
}

func (reg *Registry) Terminate() {
	if reg.publisher != nil {
		reg.publisher.Terminate()
	}
}

func (reg *Registry) PublishingRemoteComponents() bool {
	return reg.publisher != nil
}

func (reg *Registry) OnComponentChange() tools.CallbackRegistration[*ComponentChange] {
	return reg.onComponentChange
}

func (reg *Registry) OnPluginChange() tools.CallbackRegistration[*PluginChange] {
	return reg.onPluginChange
}

func (reg *Registry) AddPlugin(instanceName string, plugin *metadata.Plugin) {
	reg.lock.Lock()
	defer reg.lock.Unlock()

	key := reg.buildPluginId(instanceName, plugin)
	if _, exists := reg.pluginsPerInstance[key]; exists {
		panic(fmt.Errorf("plugin '%s' does already exist in the registry", key))
	}

	reg.pluginsPerInstance[key] = plugin
	reg.updateInstance(instanceName, func(data *instanceData) {
		data.plugins[plugin] = struct{}{}
	})

	logger.Debugf("Plugin '%s' added", key)

	reg.onPluginChange.Execute(&PluginChange{
		action:       RegistryAdd,
		instanceName: instanceName,
		plugin:       plugin,
	})
}

func (reg *Registry) RemovePlugin(instanceName string, plugin *metadata.Plugin) {
	reg.lock.Lock()
	defer reg.lock.Unlock()

	key := reg.buildPluginId(instanceName, plugin)
	if _, exists := reg.pluginsPerInstance[key]; !exists {
		panic(fmt.Errorf("plugin '%s' does not exist in the registry", key))
	}

	delete(reg.pluginsPerInstance, key)
	reg.updateInstance(instanceName, func(data *instanceData) {
		delete(data.plugins, plugin)
	})

	logger.Debugf("Plugin '%s' removed", key)

	reg.onPluginChange.Execute(&PluginChange{
		action:       RegistryRemove,
		instanceName: instanceName,
		plugin:       plugin,
	})
}

func (reg *Registry) buildPluginId(instanceName string, plugin *metadata.Plugin) string {
	if instanceName == "" {
		instanceName = "local"
	}

	return instanceName + ":" + plugin.Id()
}

func (reg *Registry) updateInstance(instanceName string, callback func(*instanceData)) {
	data := reg.instances[instanceName]
	if data == nil {
		data = &instanceData{
			plugins:    make(map[*metadata.Plugin]struct{}),
			components: make(map[Component]struct{}),
		}

		reg.instances[instanceName] = data
	}

	callback(data)

	if len(data.plugins) == 0 && len(data.components) == 0 {
		delete(reg.instances, instanceName)
	}
}

func (reg *Registry) HasPlugin(instanceName string, id string) bool {
	return reg.GetPlugin(instanceName, id) != nil
}

func (reg *Registry) GetPlugin(instanceName string, id string) *metadata.Plugin {
	reg.lock.RLock()
	defer reg.lock.RUnlock()

	if instanceName == "" {
		instanceName = "local"
	}

	key := instanceName + ":" + id
	return reg.pluginsPerInstance[key]
}

func (reg *Registry) GetPlugins(instanceName string) tools.ReadonlySlice[*metadata.Plugin] {
	reg.lock.RLock()
	defer reg.lock.RUnlock()

	data := reg.instances[instanceName]
	plugins := make([]*metadata.Plugin, 0)
	if data != nil {
		for plugin := range data.plugins {
			plugins = append(plugins, plugin)
		}
	}

	return tools.NewReadonlySlice[*metadata.Plugin](plugins)
}

func (reg *Registry) AddComponent(instanceName string, component Component) {
	reg.lock.Lock()
	defer reg.lock.Unlock()

	id := component.Id()
	if _, exists := reg.components[id]; exists {
		panic(fmt.Errorf("Component '%s' does already exist in the registry", id))
	}

	reg.components[id] = &ComponentData{
		instanceName: instanceName,
		component:    component,
	}

	reg.updateInstance(instanceName, func(data *instanceData) {
		data.components[component] = struct{}{}
	})

	logger.Debugf("Component '%s:%s' added", instanceName, id)
	reg.onComponentChange.Execute(&ComponentChange{
		action:       RegistryAdd,
		instanceName: instanceName,
		component:    component,
	})
}

func (reg *Registry) RemoveComponent(instanceName string, component Component) {
	reg.lock.Lock()
	defer reg.lock.Unlock()

	id := component.Id()
	if _, exists := reg.components[id]; !exists {
		panic(fmt.Errorf("Component '%s' does not exist in the registry", id))
	}

	delete(reg.components, id)
	reg.updateInstance(instanceName, func(data *instanceData) {
		delete(data.components, component)
	})

	logger.Debugf("Component '%s:%s' removed", instanceName, id)
	reg.onComponentChange.Execute(&ComponentChange{
		action:       RegistryRemove,
		instanceName: instanceName,
		component:    component,
	})
}

func (reg *Registry) HasComponent(id string) bool {
	return reg.GetComponentData(id) != nil
}

func (reg *Registry) GetComponent(id string) Component {
	data := reg.GetComponentData(id)
	if data == nil {
		return nil
	} else {
		return data.component
	}
}

func (reg *Registry) GetComponentData(id string) *ComponentData {
	reg.lock.RLock()
	defer reg.lock.RUnlock()

	return reg.components[id]
}

func (reg *Registry) GetComponentsData() tools.ReadonlySlice[*ComponentData] {
	reg.lock.RLock()
	defer reg.lock.RUnlock()

	components := make([]*ComponentData, 0, len(reg.components))

	for _, data := range reg.components {
		components = append(components, data)
	}

	return tools.NewReadonlySlice[*ComponentData](components)
}

func (reg *Registry) GetComponents() tools.ReadonlySlice[Component] {
	reg.lock.RLock()
	defer reg.lock.RUnlock()

	components := make([]Component, 0, len(reg.components))

	for _, data := range reg.components {
		components = append(components, data.component)
	}

	return tools.NewReadonlySlice[Component](components)
}

func (reg *Registry) GetInstanceNames() tools.ReadonlySlice[string] {
	reg.lock.RLock()
	defer reg.lock.RUnlock()

	instances := make([]string, 0, len(reg.instances))

	for instanceName := range reg.instances {
		instances = append(instances, instanceName)
	}

	return tools.NewReadonlySlice[string](instances)
}
