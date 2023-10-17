package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type BinaryToPercent struct {
	// @Config()
	Low int64

	// @Config()
	High int64

	// @State(type=range[0;100])
	Value definitions.State[int64]
}

func (component *BinaryToPercent) Init() error {
	if component.Low < 0 {
		component.Low = 0
	}

	if component.High > 100 {
		component.High = 100
	}

	component.Value.Set(component.Low)
	return nil
}

func (component *BinaryToPercent) Terminate() {
	// Noop
}

// @Action()
func (component *BinaryToPercent) SetValue(arg bool) {
	if arg {
		component.Value.Set(component.High)
	} else {
		component.Value.Set(component.Low)
	}
}
