package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ConstantPercent struct {
	// @Config(name="value")
	ConfigValue int64

	// @State(type=range[0;100])
	Value definitions.State[int64]
}

func (component *ConstantPercent) Init() error {
	value := component.ConfigValue
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}

	component.Value.Set(value)

	return nil
}

func (component *ConstantPercent) Terminate() {
	// Noop
}
