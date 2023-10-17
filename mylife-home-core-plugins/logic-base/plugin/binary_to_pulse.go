package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type BinaryToPulse struct {

	// @State()
	Off definitions.State[bool]

	// @State()
	On definitions.State[bool]
}

func (component *BinaryToPulse) Init() error {
	return nil
}

func (component *BinaryToPulse) Terminate() {
	// Noop
}

// @Action()
func (component *BinaryToPulse) Action(arg bool) {
	if arg {
		component.On.Set(true)
		component.On.Set(false)
	} else {
		component.Off.Set(true)
		component.Off.Set(false)
	}
}
