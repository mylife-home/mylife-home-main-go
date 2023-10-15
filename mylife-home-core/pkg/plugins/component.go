package plugins

import (
	"mylife-home-common/components"
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
)

var _ components.Component = (*Component)(nil)

type Component struct {
	id            string
	plugin        *Plugin
	target        definitions.Plugin
	state         map[string]untypedState
	actions       map[string]func(any)
	onStateChange *tools.CallbackManager[*components.StateChange]
}

func newComponent(id string, plugin *Plugin, target definitions.Plugin, state map[string]untypedState, actions map[string]func(any)) *Component {
	comp := &Component{
		id:            id,
		plugin:        plugin,
		target:        target,
		state:         state,
		actions:       actions,
		onStateChange: tools.NewCallbackManager[*components.StateChange](),
	}

	for name, stateItem := range comp.state {
		name := name // fix for/closure issue
		stateItem.SetOnChange(func(value any) {
			comp.onStateChange.Execute(components.NewStateChange(name, value))
		})
	}

	return comp
}

func (comp *Component) OnStateChange() tools.CallbackRegistration[*components.StateChange] {
	return comp.onStateChange
}

func (comp *Component) Id() string {
	return comp.id
}

func (comp *Component) Plugin() *metadata.Plugin {
	return comp.plugin.Metadata()
}

func (comp *Component) GetStateItem(name string) any {
	return comp.state[name].UntypedGet()
}

func (comp *Component) GetState() tools.ReadonlyMap[string, any] {
	state := make(map[string]any)

	for name, stateItem := range comp.state {
		state[name] = stateItem.UntypedGet()
	}

	return tools.NewReadonlyMap(state)
}

func (comp *Component) ExecuteAction(name string, value any) {
	action := comp.actions[name]
	action(value)
}

func (comp *Component) Terminate() {
	comp.target.Terminate()
}
