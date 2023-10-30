package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type SwitchToButton struct {
	switch_ bool

	// @State()
	Value definitions.State[bool]
}

func (component *SwitchToButton) Init(runtime definitions.Runtime) error {
	return nil
}

func (component *SwitchToButton) Terminate() {
	// Noop
}

// @Action()
func (component *SwitchToButton) Action(arg bool) {
	if component.switch_ == arg {
		return
	}

	component.switch_ = arg

	component.Value.Set(true)
	component.Value.Set(false)
}
