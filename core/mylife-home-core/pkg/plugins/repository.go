package plugins

import (
	"mylife-home-core-library/registry"

	"golang.org/x/exp/maps"
)

var plugins map[string]*Plugin

func Build() {
	plugins = make(map[string]*Plugin)

	for index := 0; index < registry.NumPlugins(); index += 1 {
		pluginType := registry.GetPlugin(index)
		plugin := buildPlugin(pluginType)
		plugins[plugin.Metadata().Id()] = plugin
	}
}

func Ids() []string {
	return maps.Keys(plugins)
}

func GetPlugin(id string) *Plugin {
	return plugins[id]
}
