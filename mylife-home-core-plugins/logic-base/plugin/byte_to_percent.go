package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ByteToPercent struct {

	// @State(type="range[0;100]")
	Value definitions.State[int64]
}

func (component *ByteToPercent) Init(runtime definitions.Runtime) error {
	return nil
}

func (component *ByteToPercent) Terminate() {
	// Noop
}

// @Action(type="range[0;255]")
func (component *ByteToPercent) Set(arg int64) {
	component.Value.Set(arg * 100 / 255)
}
