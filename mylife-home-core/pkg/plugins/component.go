package plugins

import (
	"maps"
	"mylife-home-common/components"
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
)

var _ components.Component = (*Component)(nil)

type Component struct {
	// metadata/direct component management
	id      string
	plugin  *Plugin
	target  definitions.Plugin
	state   map[string]tools.ObservableValue[any]
	actions map[string]chan any
}

type actionDispatch struct {
	name  string
	value any
}

func newComponent(id string, plugin *Plugin, target definitions.Plugin, actions map[string]func(any), state map[string]tools.ObservableValue[any]) *Component {
	// for stability since it's kept by dispatcher
	actions = maps.Clone(actions)

	comp := &Component{
		id:      id,
		plugin:  plugin,
		target:  target,
		actions: make(map[string]chan any),
		state:   state,
	}

	// make actions dispatch sequentially:
	// create one channel as unique action receiver, then dispatch
	ch := comp.initActionMerger(actions)
	go comp.dispatcher(ch, actions)

	return comp
}

func (comp *Component) initActionMerger(actions map[string]func(any)) <-chan actionDispatch {
	dummy := make(chan actionDispatch)
	merger := tools.MakeChannelMerger(dummy)

	for name := range actions {
		input := make(chan any)
		comp.actions[name] = input

		name := name
		merger.Add(tools.MapChannel(input, func(value any) actionDispatch {
			return actionDispatch{
				name:  name,
				value: value,
			}
		}))
	}

	close(dummy)

	return merger.Out()
}

func (comp *Component) dispatcher(input <-chan actionDispatch, actions map[string]func(any)) {
	for ad := range input {
		action := actions[ad.name]
		action(ad.value)
	}
}

func (comp *Component) Id() string {
	return comp.id
}

func (comp *Component) Plugin() *metadata.Plugin {
	return comp.plugin.Metadata()
}

func (comp *Component) StateItem(name string) tools.ObservableValue[any] {
	return comp.state[name]
}

func (comp *Component) Action(name string) chan<- any {
	return comp.actions[name]
}

func (comp *Component) Init() error {
	rt := makeRuntime(comp.id)
	return comp.target.Init(rt)
}

func (comp *Component) Terminate() {
	comp.target.Terminate()

	// Do not permit actions after terminate + properly close channels handlers
	for _, ch := range comp.actions {
		close(ch)
	}
}
