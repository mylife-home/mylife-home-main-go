package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ModeSelector struct {

	// @Config()
	TemperatureThreshold int64

	// @State(type="enum{cool,dry,fan-only,heat,heat-cool,off}")
	Mode definitions.State[string]

	// @State()
	Active definitions.State[bool]

	// @State(type="range[17;30]")
	Temperature definitions.State[int64]
}

func (component *ModeSelector) Init() error {
	component.Temperature.Set(17)
	component.computeMode()
	return nil
}

func (component *ModeSelector) Terminate() {
	// Noop
}

// @Action(type="range[17;30]")
func (component *ModeSelector) SetTemperature(arg int64) {
	component.Temperature.Set(arg)
	component.computeMode()
}

// @Action(type="range[17;30]")
func (component *ModeSelector) SetActive(arg bool) {
	component.Active.Set(arg)
	component.computeMode()
}

func (component *ModeSelector) computeMode() {
	if !component.Active.Get() {
		component.Mode.Set("off")
		return
	}

	if component.Temperature.Get() > component.TemperatureThreshold {
		component.Mode.Set("cool")
	} else {
		component.Mode.Set("heat")
	}
}
