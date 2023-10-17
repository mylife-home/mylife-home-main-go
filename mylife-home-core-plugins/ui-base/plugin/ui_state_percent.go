package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="ui")
type UiStatePercent struct {

	// @State(type="range[0;100]")
	Value definitions.State[int64]
}

func (component *UiStatePercent) Init() error {
	return nil
}

func (component *UiStatePercent) Terminate() {
	// Noop
}

// @Action(type="range[0;100]")
func (component *UiStatePercent) SetValue(arg int64) {
	component.Value.Set(arg)
}
