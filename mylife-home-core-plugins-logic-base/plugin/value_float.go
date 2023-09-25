package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(description="binary value description" usage="logic" version="1.0.0")
type ValueFloat struct {
	// @Config(description="initial value")
	InitialValue float64

	// @State(description="current value")
	Value definitions.State[float64]
}

// @Action(description="set current value")
func (component *ValueFloat) SetValue(arg float64) {
	component.Value.Set(arg)
}

func (component *ValueFloat) Init() error {
	component.Value.Set(component.InitialValue)
	return nil
}

func (component *ValueFloat) Terminate() {
	// Noop
}
