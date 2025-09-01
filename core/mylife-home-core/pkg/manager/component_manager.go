package manager

import (
	"encoding/json"
	"fmt"
	"mylife-home-common/components"
	"mylife-home-common/components/metadata"
	"mylife-home-common/instance_info"
	"mylife-home-core/pkg/plugins"
	"mylife-home-core/pkg/store"
	"strings"

	"github.com/gookit/goutil/errorx/panics"
	"golang.org/x/exp/slices"
)

type componentManager struct {
	registry         components.Registry
	store            *store.Store
	supportsBindings bool
	components       map[string]*plugins.Component
	bindings         map[string]*binding
}

func makeComponentManager(registry components.Registry, supportsBindings bool) *componentManager {

	manager := &componentManager{
		registry:         registry,
		store:            store.MakeStore(),
		supportsBindings: supportsBindings,
		components:       make(map[string]*plugins.Component),
		bindings:         make(map[string]*binding),
	}

	instance_info.AddCapability("components-manager")
	if manager.supportsBindings {
		instance_info.AddCapability("bindings-manager")
	}

	for _, id := range plugins.Ids() {
		pluginInstance := plugins.GetPlugin(id)
		manager.registry.AddPlugin("", pluginInstance.Metadata())
	}

	err := manager.store.Load()
	panics.IsTrue(err == nil, "error loading store: %s", err)

	if !manager.supportsBindings && manager.store.HasBindings() {
		panic("store has bindings but configuration does not activate its support")
	}

	for _, config := range manager.store.GetComponents() {
		pluginInstance := plugins.GetPlugin(config.Plugin)
		panics.IsTrue(pluginInstance != nil, "plugin does not exists: '%s'", config.Plugin)

		pluginConfig, err := manager.buildConfig(pluginInstance.Metadata(), config.Config)
		panics.IsTrue(err == nil, "could not create plugin config for plugin '%s': %s", pluginInstance.Metadata().Id(), err)

		comp, err := pluginInstance.Instantiate(config.Id, pluginConfig)
		panics.IsTrue(err == nil, "could not create component '%s' (plugin='%s'): %s", config.Id, pluginInstance.Metadata().Id(), err)

		manager.components[config.Id] = comp
		manager.registry.AddComponent("", comp)
	}

	for _, config := range manager.store.GetBindings() {
		key := manager.buildBindingKey(config)
		manager.bindings[key] = makeBinding(manager.registry, config)
	}

	return manager
}

func (manager *componentManager) Terminate() {
	for _, binding := range manager.bindings {
		binding.Terminate()
	}
	clear(manager.bindings)

	for _, component := range manager.components {
		manager.registry.RemoveComponent("", component)
		component.Terminate()
	}
	clear(manager.components)
}

func (manager *componentManager) AddComponent(id string, plugin string, config map[string]json.RawMessage) error {
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
	return slices.Clone(manager.store.GetComponents())
}

func (manager *componentManager) AddBinding(config *store.BindingConfig) error {
	key := manager.buildBindingKey(config)
	if _, exists := manager.bindings[key]; exists {
		return fmt.Errorf("binding already exists: '%s'", config)
	}

	manager.bindings[key] = makeBinding(manager.registry, config)
	manager.store.AddBinding(config)

	return nil
}

func (manager *componentManager) RemoveBinding(config *store.BindingConfig) error {
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
	return slices.Clone(manager.store.GetBindings())
}

func (manager *componentManager) Save() error {
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
