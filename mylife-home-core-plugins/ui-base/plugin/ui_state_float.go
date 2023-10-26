package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="ui")
type UiStateFloat struct {

	// @State()
	Value definitions.State[float64]
}

func (component *UiStateFloat) Init(runtime definitions.Runtime) error {
	return nil
}

func (component *UiStateFloat) Terminate() {
	// Noop
}

// @Action()
func (component *UiStateFloat) SetValue(arg float64) {
	component.Value.Set(arg)
}
