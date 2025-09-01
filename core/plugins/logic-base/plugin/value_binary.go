package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ValueBinary struct {
	// @State()
	Value definitions.State[bool]
}

// @Action()
func (component *ValueBinary) SetValue(arg bool) {
	component.Value.Set(arg)
}

func (component *ValueBinary) Init(runtime definitions.Runtime) error {
	return nil
}

func (component *ValueBinary) Terminate() {
	// Noop
}
