package registry

import (
	"mylife-home-common/components/metadata"
	"reflect"
)

type PluginType struct {
	target  reflect.Type
	meta    *metadata.Plugin
	state   []*StateType
	actions []*ActionType
	config  []*ConfigType
}

func (plugin *PluginType) Target() reflect.Type {
	return plugin.target
}

func (plugin *PluginType) Metadata() *metadata.Plugin {
	return plugin.meta
}

func (plugin *PluginType) NumState() int {
	return len(plugin.state)
}

func (plugin *PluginType) StateItem(index int) *StateType {
	return plugin.state[index]
}

func (plugin *PluginType) NumActions() int {
	return len(plugin.actions)
}

func (plugin *PluginType) Action(index int) *ActionType {
	return plugin.actions[index]
}

func (plugin *PluginType) NumConfig() int {
	return len(plugin.config)
}

func (plugin *PluginType) ConfigItem(index int) *ConfigType {
	return plugin.config[index]
}

type StateType struct {
	target *reflect.StructField
	meta   *metadata.Member
}

func (state *StateType) Target() *reflect.StructField {
	return state.target
}

func (state *StateType) Metadata() *metadata.Member {
	return state.meta
}

type ActionType struct {
	target *reflect.Method
	meta   *metadata.Member
}

func (action *ActionType) Target() *reflect.Method {
	return action.target
}

func (action *ActionType) Metadata() *metadata.Member {
	return action.meta
}

type ConfigType struct {
	target *reflect.StructField
	meta   *metadata.ConfigItem
}

func (config *ConfigType) Target() *reflect.StructField {
	return config.target
}

func (config *ConfigType) Metadata() *metadata.ConfigItem {
	return config.meta
}
