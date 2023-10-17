package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="ui")
type UiStateNullablePercent struct {

	// @State(type="range[-1;100]")
	Value definitions.State[int64]
}

func (component *UiStateNullablePercent) Init() error {
	return nil
}

func (component *UiStateNullablePercent) Terminate() {
	// Noop
}

// @Action(type="range[-1;100]")
func (component *UiStateNullablePercent) SetValue(arg int64) {
	component.Value.Set(arg)
}
