package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type PercentToBinary struct {
	// @Config()
	Threshold int64

	// @State()
	Value definitions.State[bool]
}

func (component *PercentToBinary) Init(runtime definitions.Runtime) error {
	return nil
}

func (component *PercentToBinary) Terminate() {
	// Noop
}

// @Action(type="range[0;100]")
func (component *PercentToBinary) SetValue(arg int64) {
	component.Value.Set(arg >= component.Threshold)
}
