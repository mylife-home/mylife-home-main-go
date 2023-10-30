package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type TemperatureSelector struct {

	// @Config()
	Initial int64

	// @State(type="range[17;30]")
	Value definitions.State[int64]
}

func (component *TemperatureSelector) Init(runtime definitions.Runtime) error {
	component.Value.Set(component.Initial)

	return nil
}

func (component *TemperatureSelector) Terminate() {
	// Noop
}

// @Action()
func (component *TemperatureSelector) Up(arg bool) {
	if arg {
		newValue := component.Value.Get() + 1
		if newValue > 30 {
			newValue = 30
		}

		component.Value.Set(newValue)
	}
}

// @Action()
func (component *TemperatureSelector) Down(arg bool) {
	if arg {
		newValue := component.Value.Get() - 1
		if newValue < 17 {
			newValue = 17
		}

		component.Value.Set(newValue)
	}
}
