package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="ui")
type UiStateBool struct {

	// @State()
	Value definitions.State[bool]
}

func (component *UiStateBool) Init() error {
	return nil
}

func (component *UiStateBool) Terminate() {
	// Noop
}

// @Action()
func (component *UiStateBool) SetValue(arg bool) {
	component.Value.Set(arg)
}
