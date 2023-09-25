package plugins

import "mylife-home-core-library/definitions"

type Component struct {
	id      string
	plugin  *Plugin
	target  definitions.Plugin
	state   map[string]untypedState
	actions map[string]func(any)
}

func (comp *Component) Id() string {
	return comp.id
}

func (comp *Component) Plugin() *Plugin {
	return comp.plugin
}

func (comp *Component) SetOnStateChange(callback func(name string, value any)) {
	for name, stateItem := range comp.state {
		stateItem.SetOnChange(func(value any) {
			callback(name, value)
		})
	}
}

func (comp *Component) GetStateItem(name string) any {
	return comp.state[name].UntypedGet()
}

func (comp *Component) GetState() map[string]any {
	state := make(map[string]any)

	for name, stateItem := range comp.state {
		state[name] = stateItem.UntypedGet()
	}

	return state
}

func (comp *Component) Execute(name string, value any) {
	action := comp.actions[name]
	action(value)
}

func (comp *Component) Termainte() {
	comp.target.Terminate()
}
