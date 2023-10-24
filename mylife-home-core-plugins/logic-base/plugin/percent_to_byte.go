package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type PercentToByte struct {

	// @State(type="range[0;255]")
	Value definitions.State[int64]
}

func (component *PercentToByte) Init() error {
	return nil
}

func (component *PercentToByte) Terminate() {
	// Noop
}

// @Action(type="range[0;100]")
func (component *PercentToByte) SetValue(arg int64) {
	component.Value.Set(arg * 255 / 100)
}
