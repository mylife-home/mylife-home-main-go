package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="ui")
type UiStateText struct {

	// @State()
	Value definitions.State[string]
}

func (component *UiStateText) Init(runtime definitions.Runtime) error {
	return nil
}

func (component *UiStateText) Terminate() {
	// Noop
}

// @Action()
func (component *UiStateText) SetValue(arg string) {
	component.Value.Set(arg)
}
