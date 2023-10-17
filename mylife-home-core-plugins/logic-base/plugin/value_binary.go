package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(description="binary value description" usage="logic" version="1.0.0")
type ValueBinary struct {
	// @Config(description="initial value")
	InitialValue bool

	// @State(description="current value")
	Value definitions.State[bool]
}

// @Action(description="set current value")
func (component *ValueBinary) SetValue(arg bool) {
	component.Value.Set(arg)
}

func (component *ValueBinary) Init() error {
	component.Value.Set(component.InitialValue)
	return nil
}

func (component *ValueBinary) Terminate() {
	// Noop
}
