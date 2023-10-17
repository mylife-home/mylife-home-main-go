package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type StepRelay struct {

	// @State()
	Value definitions.State[bool]
}

func (component *StepRelay) Init() error {
	return nil
}

func (component *StepRelay) Terminate() {
	// Noop
}

// @Action()
func (component *StepRelay) Action(arg bool) {
	if arg {
		component.Value.Set(!component.Value.Get())

	}
}

// @Action()
func (component *StepRelay) On(arg bool) {
	if arg {
		component.Value.Set(true)

	}
}

// @Action()
func (component *StepRelay) Off(arg bool) {
	if arg {
		component.Value.Set(false)
	}
}
