package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type TemperatureToFloat struct {

	// @State()
	Value definitions.State[float64]
}

func (component *TemperatureToFloat) Init(runtime definitions.Runtime) error {
	return nil
}

func (component *TemperatureToFloat) Terminate() {
	// Noop
}

// @Action(type="range[17;30]")
func (component *TemperatureToFloat) SetValue(arg int64) {
	component.Value.Set(float64(arg))
}
