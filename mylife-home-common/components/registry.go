package components

import (
	"fmt"
	"mylife-home-common/components/metadata"
	"mylife-home-common/log"
	"mylife-home-common/tools"
	"sync"

	"github.com/gookit/goutil/errorx/panics"
	"golang.org/x/exp/maps"
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

type ComponentData interface {
	InstanceName() string
	Component() Component
}

type Registry interface {
	OnComponentChange() tools.Observable[*ComponentChange]
	OnPluginChange() tools.Observable[*PluginChange]

	AddPlugin(instanceName string, plugin *metadata.Plugin)
	RemovePlugin(instanceName string, plugin *metadata.Plugin)
	HasPlugin(instanceName string, id string) bool
	GetPlugin(instanceName string, id string) *metadata.Plugin
	GetPlugins(instanceName string) []*metadata.Plugin

	AddComponent(instanceName string, component Component)
	RemoveComponent(instanceName string, component Component)
	HasComponent(id string) bool
	GetComponent(id string) Component
	GetComponentData(id string) ComponentData
	GetComponentsData() []ComponentData
	GetComponents() []Component

	GetInstanceNames() []string
}

type instanceData struct {
	plugins    map[string]*metadata.Plugin
	components map[string]Component
}

var _ ComponentData = (*componentData)(nil)

type componentData struct {
	instanceName string
	component    Component
}

func (data *componentData) InstanceName() string {
	return data.instanceName
}

func (data *componentData) Component() Component {
	return data.component
}

var _ Registry = (*registry)(nil)

type registry struct {
	onComponentChange tools.Subject[*ComponentChange]
	onPluginChange    tools.Subject[*PluginChange]

	components map[string]*componentData
	instances  map[string]*instanceData

	mux sync.Mutex
}

func NewRegistry() Registry {
	return &registry{
		onComponentChange: tools.MakeSubject[*ComponentChange](),
		onPluginChange:    tools.MakeSubject[*PluginChange](),
		components:        make(map[string]*componentData),
		instances:         make(map[string]*instanceData),
	}
}

func (reg *registry) OnComponentChange() tools.Observable[*ComponentChange] {
	return reg.onComponentChange
}

func (reg *registry) OnPluginChange() tools.Observable[*PluginChange] {
	return reg.onPluginChange
}

func (reg *registry) AddPlugin(instanceName string, plugin *metadata.Plugin) {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	id := plugin.Id()
	logId := reg.buildLogId(instanceName, id)

	reg.updateInstance(instanceName, func(data *instanceData) {
		_, exists := data.plugins[id]
		panics.IsTrue(!exists, "plugin '%s' does already exist in the registry", logId)

		data.plugins[id] = plugin
	})

	logger.Debugf("Plugin '%s' added", logId)

	reg.onPluginChange.Notify(&PluginChange{
		action:       RegistryAdd,
		instanceName: instanceName,
		plugin:       plugin,
	})
}

func (reg *registry) RemovePlugin(instanceName string, plugin *metadata.Plugin) {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	id := plugin.Id()
	logId := reg.buildLogId(instanceName, id)

	reg.updateInstance(instanceName, func(data *instanceData) {
		_, exists := data.plugins[id]
		panics.IsTrue(exists, "plugin '%s' does not exist in the registry", logId)

		delete(data.plugins, id)
	})

	logger.Debugf("Plugin '%s' removed", logId)
	reg.onPluginChange.Notify(&PluginChange{
		action:       RegistryRemove,
		instanceName: instanceName,
		plugin:       plugin,
	})
}

func (reg *registry) updateInstance(instanceName string, callback func(*instanceData)) {
	data := reg.instances[instanceName]
	if data == nil {
		data = &instanceData{
			plugins:    make(map[string]*metadata.Plugin),
			components: make(map[string]Component),
		}

		reg.instances[instanceName] = data
		logger.Debugf("Instance '%s' added", instanceName)
	}

	callback(data)

	if len(data.plugins) == 0 && len(data.components) == 0 {
		delete(reg.instances, instanceName)
		logger.Debugf("Instance '%s' removed", instanceName)
	}
}

func (reg *registry) HasPlugin(instanceName string, id string) bool {
	return reg.GetPlugin(instanceName, id) != nil
}

func (reg *registry) GetPlugin(instanceName string, id string) *metadata.Plugin {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	instance := reg.instances[instanceName]
	if instance == nil {
		return nil
	}

	return instance.plugins[id]
}

func (reg *registry) GetPlugins(instanceName string) []*metadata.Plugin {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	instance := reg.instances[instanceName]
	if instance == nil {
		return []*metadata.Plugin{}
	}

	return maps.Values(instance.plugins)
}

func (reg *registry) AddComponent(instanceName string, component Component) {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	id := component.Id()
	logId := reg.buildLogId(instanceName, component.Id())

	if _, exists := reg.components[id]; exists {
		panic(fmt.Errorf("Component '%s' does already exist in the registry", id))
	}

	reg.components[id] = &componentData{
		instanceName: instanceName,
		component:    component,
	}

	reg.updateInstance(instanceName, func(data *instanceData) {
		data.components[id] = component
	})

	logger.Debugf("Component '%s' added", logId)

	reg.onComponentChange.Notify(&ComponentChange{
		action:       RegistryAdd,
		instanceName: instanceName,
		component:    component,
	})
}

func (reg *registry) RemoveComponent(instanceName string, component Component) {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	id := component.Id()
	logId := reg.buildLogId(instanceName, component.Id())

	if _, exists := reg.components[id]; !exists {
		panic(fmt.Errorf("Component '%s' does not exist in the registry", id))
	}

	delete(reg.components, id)
	reg.updateInstance(instanceName, func(data *instanceData) {
		delete(data.components, id)
	})

	logger.Debugf("Component '%s' removed", logId)

	reg.onComponentChange.Notify(&ComponentChange{
		action:       RegistryRemove,
		instanceName: instanceName,
		component:    component,
	})
}

func (reg *registry) HasComponent(id string) bool {
	return reg.GetComponentData(id) != nil
}

func (reg *registry) GetComponent(id string) Component {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	data := reg.components[id]
	if data == nil {
		return nil
	} else {
		return data.component
	}
}

func (reg *registry) GetComponentData(id string) ComponentData {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	compData, exists := reg.components[id]
	if exists {
		return compData
	} else {
		// Mandatory to return a nil value as interface
		return nil
	}
}

func (reg *registry) GetComponentsData() []ComponentData {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	components := make([]ComponentData, len(reg.components))

	index := 0
	for _, data := range reg.components {
		components[index] = data
		index += 1
	}

	return components
}

func (reg *registry) GetComponents() []Component {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	components := make([]Component, len(reg.components))

	index := 0
	for _, data := range reg.components {
		components[index] = data.component
		index += 1
	}

	return components
}

func (reg *registry) GetInstanceNames() []string {
	reg.mux.Lock()
	defer reg.mux.Unlock()

	return maps.Keys(reg.instances)
}

func (reg *registry) buildLogId(instanceName string, id string) string {
	if instanceName == "" {
		return "<local>:" + id
	} else {
		return instanceName + ":" + id
	}
}
