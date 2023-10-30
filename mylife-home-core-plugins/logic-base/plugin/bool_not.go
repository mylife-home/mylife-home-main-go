package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type BoolNot struct {

	// @State()
	Value definitions.State[bool]
}

func (component *BoolNot) Init(runtime definitions.Runtime) error {
	return nil
}

func (component *BoolNot) Terminate() {
	// Noop
}

// @Action()
func (component *BoolNot) Set(arg bool) {
	component.Value.Set(!arg)
}
