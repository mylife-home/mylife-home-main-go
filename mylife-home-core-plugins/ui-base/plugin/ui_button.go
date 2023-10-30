package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="ui")
type UiButton struct {

	// @State()
	Value definitions.State[bool]
}

func (component *UiButton) Init(runtime definitions.Runtime) error {
	return nil
}

func (component *UiButton) Terminate() {
	// Noop
}

// @Action()
func (component *UiButton) Action(arg bool) {
	component.Value.Set(arg)
}
