package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ValueNullablePercent struct {
	// @State(type="range[-1;100]")
	Value definitions.State[int64]
}

// @Action(type="range[-1;100]")
func (component *ValueNullablePercent) SetValue(arg int64) {
	component.Value.Set(arg)
}

func (component *ValueNullablePercent) Init(runtime definitions.Runtime) error {
	return nil
}

func (component *ValueNullablePercent) Terminate() {
	// Noop
}
