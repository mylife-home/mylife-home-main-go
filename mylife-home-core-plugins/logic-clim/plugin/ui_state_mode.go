package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="ui")
type UiStateMode struct {

	// @State(type="enum{cool,dry,fan-only,heat,heat-cool,off}")
	Value definitions.State[string]
}

func (component *UiStateMode) Init() error {
	component.Value.Set("off")

	return nil
}

func (component *UiStateMode) Terminate() {
	// Noop
}

// @Action(type="enum{cool,dry,fan-only,heat,heat-cool,off}")
func (component *UiStateMode) SetValue(arg string) {
	component.Value.Set(arg)
}
