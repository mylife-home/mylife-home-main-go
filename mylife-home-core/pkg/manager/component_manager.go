package manager

import (
	"encoding/json"
	"fmt"
	"mylife-home-common/bus"
	"mylife-home-common/components"
	"mylife-home-common/components/metadata"
	"mylife-home-common/instance_info"
	"mylife-home-core/pkg/plugins"
	"mylife-home-core/pkg/store"
	"strings"
	"sync"

	"golang.org/x/exp/slices"
)

type componentManager struct {
	transport        *bus.Transport
	supportsBindings bool
	registry         *components.Registry
	publisher        *busPublisher
	store            *store.Store
	components       map[string]*plugins.Component
	bindings         map[string]*binding
	mux              sync.Mutex
}

func makeComponentManager(transport *bus.Transport) (*componentManager, error) {
	store, err := store.MakeStore()
	if err != nil {
		return nil, err
	}

	manager := &componentManager{
		transport:  transport,
		store:      store,
		components: make(map[string]*plugins.Component),
		bindings:   make(map[string]*binding),
	}

	manager.supportsBindings = manager.transport.Presence().Tracking()

	options := components.NewRegistryOptions()
	if manager.supportsBindings {
		options.PublishRemoteComponents(manager.transport)
	}
	manager.registry = components.NewRegistry(options)

	manager.publisher = newBusPublisher(manager.transport, manager.registry)

	instance_info.AddCapability("components-manager")
	if manager.supportsBindings {
		instance_info.AddCapability("bindings-manager")
	}

	for _, id := range plugins.Ids() {
		pluginInstance := plugins.GetPlugin(id)
		manager.registry.AddPlugin("", pluginInstance.Metadata())
	}

	if err := manager.store.Load(); err != nil {
		return nil, err
	}

	if !manager.supportsBindings && manager.store.HasBindings() {
		return nil, fmt.Errorf("store has bindings but configuration does not activate its support")
	}

	for _, config := range manager.store.GetComponents() {
		pluginInstance := plugins.GetPlugin(config.Plugin)
		if pluginInstance == nil {
			return nil, fmt.Errorf("plugin does not exists: '%s'", config.Plugin)
		}

		pluginConfig, err := manager.buildConfig(pluginInstance.Metadata(), config.Config)
		if err != nil {
			return nil, fmt.Errorf("could not create plugin config for plugin '%s': %w", pluginInstance.Metadata().Id(), err)
		}

		comp, err := pluginInstance.Instantiate(config.Id, pluginConfig)
		if err != nil {
			return nil, err
		}

		manager.components[config.Id] = comp
		manager.registry.AddComponent("", comp)
	}

	for _, config := range manager.store.GetBindings() {
		key := manager.buildBindingKey(config)
		manager.bindings[key] = makeBinding(manager.registry, config)
	}

	return manager, nil
}

func (manager *componentManager) Terminate() {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	for _, binding := range manager.bindings {
		binding.Terminate()
	}
	clear(manager.bindings)

	for _, component := range manager.components {
		manager.registry.RemoveComponent("", component)
		component.Terminate()
	}
	clear(manager.components)

	manager.publisher.Terminate()
	manager.registry.Terminate()
}

func (manager *componentManager) SupportsBindings() bool {
	return manager.supportsBindings
}

func (manager *componentManager) AddComponent(id string, plugin string, config map[string]json.RawMessage) error {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	if _, exists := manager.components[id]; exists {
		return fmt.Errorf("component id duplicate: '%s'", id)
	}

	pluginInstance := plugins.GetPlugin(plugin)
	if pluginInstance == nil {
		return fmt.Errorf("plugin does not exists: '%s'", plugin)
	}

	pluginConfig, err := manager.buildConfig(pluginInstance.Metadata(), config)
	if err != nil {
		return fmt.Errorf("could not create plugin config for plugin '%s': %w", pluginInstance.Metadata().Id(), err)
	}

	comp, err := pluginInstance.Instantiate(id, pluginConfig)
	if err != nil {
		return err
	}

	manager.components[id] = comp
	manager.registry.AddComponent("", comp)
	manager.store.SetComponent(&store.ComponentConfig{
		Id:     id,
		Plugin: plugin,
		Config: config,
	})

	return nil
}

func (manager *componentManager) RemoveComponent(id string) error {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	comp, exists := manager.components[id]
	if !exists {
		return fmt.Errorf("component id does not exist: '%s'", id)
	}

	manager.registry.RemoveComponent("", comp)
	comp.Terminate()
	delete(manager.components, id)
	manager.store.RemoveComponent(id)

	return nil
}

func (manager *componentManager) GetComponents() []*store.ComponentConfig {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	return slices.Clone(manager.store.GetComponents())
}

func (manager *componentManager) AddBinding(config *store.BindingConfig) error {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	key := manager.buildBindingKey(config)
	if _, exists := manager.bindings[key]; exists {
		return fmt.Errorf("binding already exists: '%s'", config)
	}

	manager.bindings[key] = makeBinding(manager.registry, config)
	manager.store.AddBinding(config)

	return nil
}

func (manager *componentManager) RemoveBinding(config *store.BindingConfig) error {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	key := manager.buildBindingKey(config)
	binding, exists := manager.bindings[key]
	if !exists {
		return fmt.Errorf("binding does not exist: %s", config)
	}

	binding.Terminate()
	delete(manager.bindings, key)
	manager.store.RemoveBinding(config)

	return nil
}

func (manager *componentManager) GetBindings() []*store.BindingConfig {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	return slices.Clone(manager.store.GetBindings())
}

func (manager *componentManager) Save() error {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	return manager.store.Save()
}

func (manager *componentManager) buildBindingKey(config *store.BindingConfig) string {
	return strings.Join([]string{config.SourceComponent, config.SourceState, config.TargetComponent, config.TargetAction}, ":")
}

// Deserialize properly. go unmarshaller does not differentiate properly int vs float
func (manager *componentManager) buildConfig(plugin *metadata.Plugin, config map[string]json.RawMessage) (map[string]any, error) {
	result := make(map[string]any)

	for _, name := range plugin.ConfigNames() {
		item := plugin.Config(name)
		raw, ok := config[name]
		if !ok {
			return nil, fmt.Errorf("missing config value for item '%s'", name)
		}

		switch item.ValueType() {
		case metadata.String:
			{
				var value string
				if err := json.Unmarshal(raw, &value); err != nil {
					return nil, fmt.Errorf("could not read value '%s' for item '%s': %w", string(raw), name, err)
				}

				result[name] = value
			}

		case metadata.Bool:
			{
				var value bool
				if err := json.Unmarshal(raw, &value); err != nil {
					return nil, fmt.Errorf("could not read value '%s' for item '%s': %w", string(raw), name, err)
				}

				result[name] = value
			}

		case metadata.Integer:
			{
				var value int64
				if err := json.Unmarshal(raw, &value); err != nil {
					return nil, fmt.Errorf("could not read value '%s' for item '%s': %w", string(raw), name, err)
				}

				result[name] = value
			}

		case metadata.Float:
			{
				var value float64
				if err := json.Unmarshal(raw, &value); err != nil {
					return nil, fmt.Errorf("could not read value '%s' for item '%s': %w", string(raw), name, err)
				}

				result[name] = value
			}

		default:
			return nil, fmt.Errorf("unhandled value type '%s' for item '%s'", item.ValueType(), name)
		}
	}

	return result, nil
}
