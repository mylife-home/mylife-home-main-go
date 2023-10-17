package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ConstantByte struct {
	// @Config(name="value")
	ConfigValue int64

	// @State(type="range[0;255]")
	Value definitions.State[int64]
}

func (component *ConstantByte) Init() error {
	value := component.ConfigValue
	if value < 0 {
		value = 0
	}
	if value > 255 {
		value = 255
	}

	component.Value.Set(value)

	return nil
}

func (component *ConstantByte) Terminate() {
	// Noop
}
