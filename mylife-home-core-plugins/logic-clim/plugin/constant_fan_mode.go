package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ConstantFanMode struct {

	// @Config(name="value" description="One of 'auto', 'high', 'low', 'medium', 'quiet'")
	ConfigValue string

	// @State(type="enum{auto,high,low,medium,quiet}")
	Value definitions.State[string]
}

func (component *ConstantFanMode) Init(runtime definitions.Runtime) error {
	component.Value.Set(component.ConfigValue)

	return nil
}

func (component *ConstantFanMode) Terminate() {
	// Noop
}
