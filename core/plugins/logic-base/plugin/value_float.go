package plugin

import (
	"math"
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ValueFloat struct {
	// @State()
	Value definitions.State[float64]
}

// @Action()
func (component *ValueFloat) SetValue(arg float64) {
	component.Value.Set(arg)
}

func (component *ValueFloat) Init(runtime definitions.Runtime) error {
	component.Value.Set(math.NaN())
	return nil
}

func (component *ValueFloat) Terminate() {
	// Noop
}
