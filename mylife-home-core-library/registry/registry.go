package registry

var plugins []*PluginType = make([]*PluginType, 0)

func RegisterPlugin(plugin *PluginType) {
	plugins = append(plugins, plugin)
}

func NumPlugins() int {
	return len(plugins)
}

func GetPlugin(index int) *PluginType {
	return plugins[index]
}
