package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ConstantBool struct {
	// @Config(name="value")
	ConfigValue bool

	// @State()
	Value definitions.State[bool]
}

func (component *ConstantBool) Init(runtime definitions.Runtime) error {
	component.Value.Set(component.ConfigValue)

	return nil
}

func (component *ConstantBool) Terminate() {
	// Noop
}
