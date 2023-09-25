package plugins

import (
	"fmt"
	"mylife-home-common/components/metadata"
	"mylife-home-common/log"
	"mylife-home-core-library/definitions"
	"mylife-home-core-library/registry"
	"reflect"
)

var logger = log.CreateLogger("mylife:home:core:plugins")

type Plugin struct {
	target  reflect.Type
	meta    *metadata.Plugin
	state   map[string]*pluginStateItem
	actions map[string]*pluginAction
	config  map[string]*pluginConfigItem
}

func buildPlugin(pluginType *registry.PluginType) *Plugin {
	plugin := &Plugin{
		target:  pluginType.Target(),
		meta:    pluginType.Metadata(),
		state:   make(map[string]*pluginStateItem),
		actions: make(map[string]*pluginAction),
		config:  make(map[string]*pluginConfigItem),
	}

	for i := 0; i < pluginType.NumState(); i += 1 {
		stateType := pluginType.StateItem(i)
		name := stateType.Metadata().Name()
		plugin.state[name] = makeStateItem(stateType)
	}

	for i := 0; i < pluginType.NumActions(); i += 1 {
		actionType := pluginType.Action(i)
		name := actionType.Metadata().Name()
		plugin.actions[name] = makeAction(actionType)
	}

	for i := 0; i < pluginType.NumConfig(); i += 1 {
		configType := pluginType.ConfigItem(i)
		name := configType.Metadata().Name()
		plugin.config[name] = makeConfigItem(configType)
	}

	logger.Infof("Plugin loaded '%s'", plugin.meta.Id())

	return plugin
}

func (plugin *Plugin) Metadata() *metadata.Plugin {
	return plugin.meta
}

func (plugin *Plugin) Instantiate(id string, config map[string]any) (*Component, error) {
	if err := plugin.validateConfig(config); err != nil {
		return nil, err
	}

	// Create instance
	compPtr := reflect.New(plugin.target)

	// Prepare it
	state := make(map[string]untypedState)
	actions := make(map[string]func(any))

	for name, stateItem := range plugin.state {
		state[name] = stateItem.init(compPtr)
	}

	for name, action := range plugin.actions {
		actions[name] = action.init(compPtr)
	}

	for name, configItem := range plugin.config {
		// Note: already checked in validateConfig
		value := config[name]
		configItem.configure(compPtr, value)
	}

	target := compPtr.Interface().(definitions.Plugin)

	// Initialize the component
	if err := target.Init(); err != nil {
		return nil, err
	}

	comp := &Component{
		id:      id,
		plugin:  plugin,
		target:  target,
		state:   state,
		actions: actions,
	}

	logger.Infof("Component created: '%s'", comp.id)
	logger.Debugf("Configuration applied (component='%s'): %+v", comp.id, config)

	return comp, nil
}

func (plugin *Plugin) validateConfig(config map[string]any) error {
	for name, item := range plugin.config {
		value, ok := config[name]
		if !ok {
			return fmt.Errorf("missing value for configuration '%s'", name)
		}

		if err := item.validate(value); err != nil {
			return err
		}
	}

	return nil
}

type pluginStateItem struct {
	target *reflect.StructField
	meta   *metadata.Member
}

func makeStateItem(stateType *registry.StateType) *pluginStateItem {
	return &pluginStateItem{
		target: stateType.Target(),
		meta:   stateType.Metadata(),
	}
}

func (s *pluginStateItem) init(compPtr reflect.Value) untypedState {
	impl := makeStateImpl(s.meta.ValueType())

	comp := compPtr.Elem()
	comp.FieldByName(s.target.Name).Set(reflect.ValueOf(impl))

	return impl
}

type pluginAction struct {
	target *reflect.Method
	meta   *metadata.Member
}

func makeAction(actionType *registry.ActionType) *pluginAction {
	return &pluginAction{
		target: actionType.Target(),
		meta:   actionType.Metadata(),
	}
}

func (a *pluginAction) init(compPtr reflect.Value) func(any) {
	fn := a.target.Func

	return func(arg any) {
		fn.Call([]reflect.Value{compPtr, reflect.ValueOf(arg)})
	}
}

type pluginConfigItem struct {
	target *reflect.StructField
	meta   *metadata.ConfigItem
}

func makeConfigItem(configType *registry.ConfigType) *pluginConfigItem {
	return &pluginConfigItem{
		target: configType.Target(),
		meta:   configType.Metadata(),
	}
}

func (c *pluginConfigItem) configure(compPtr reflect.Value, value any) {
	comp := compPtr.Elem()
	comp.FieldByName(c.target.Name).Set(reflect.ValueOf(value))
}

func (c *pluginConfigItem) validate(value any) error {
	if !c.meta.ValueType().Validate(value) {
		return fmt.Errorf("invalid value '%v' for configuration '%s'", value, c.meta.Name())
	}

	return nil
}
