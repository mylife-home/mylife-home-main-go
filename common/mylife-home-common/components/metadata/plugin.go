package metadata

import (
	"fmt"
	"strings"

	"golang.org/x/exp/maps"
)

type PluginUsage string

const (
	Sensor   PluginUsage = "sensor"
	Actuator PluginUsage = "actuator"
	Logic    PluginUsage = "logic"
	Ui       PluginUsage = "ui"
)

type Plugin struct {
	module      string
	name        string
	description string
	usage       PluginUsage
	version     string
	config      map[string]*ConfigItem
	members     map[string]*Member
}

func (plugin *Plugin) Module() string {
	return plugin.module
}

func (plugin *Plugin) Name() string {
	return plugin.name
}

func (plugin *Plugin) Id() string {
	return plugin.module + "." + plugin.name
}

func (plugin *Plugin) Description() string {
	return plugin.description
}

func (plugin *Plugin) Usage() PluginUsage {
	return plugin.usage
}

func (plugin *Plugin) Version() string {
	return plugin.version
}

func (plugin *Plugin) ConfigNames() []string {
	return maps.Keys(plugin.config)
}

func (plugin *Plugin) Config(name string) *ConfigItem {
	return plugin.config[name]
}

func (plugin *Plugin) MemberNames() []string {
	return maps.Keys(plugin.members)
}

func (plugin *Plugin) Member(name string) *Member {
	return plugin.members[name]
}

func (plugin *Plugin) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s.%s, usage=%s, version=%s, members=[", plugin.module, plugin.name, plugin.usage, plugin.version))

	var first = true
	for name, member := range plugin.members {
		if first {
			first = false
		} else {
			builder.WriteString(", ")
		}

		builder.WriteString(fmt.Sprintf("%s(%s %s)", name, member.memberType, member.valueType))
	}

	builder.WriteString("], config=[")

	first = true
	for name, configItem := range plugin.config {
		if first {
			first = false
		} else {
			builder.WriteString(", ")
		}

		builder.WriteString(fmt.Sprintf("%s(%s)", name, configItem.valueType))
	}

	builder.WriteString("]")

	return builder.String()
}
