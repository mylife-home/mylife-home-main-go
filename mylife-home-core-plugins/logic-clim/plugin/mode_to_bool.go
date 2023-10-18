package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ModeToBool struct {

	// @Config(name="trueValue" description="One of 'cool', 'dry', 'fan-only', 'heat', 'heat-cool', 'off'")
	TrueValue string

	// @State()
	Value definitions.State[bool]
}

func (component *ModeToBool) Init() error {
	return nil
}

func (component *ModeToBool) Terminate() {
	// Noop
}

// @Action(type="enum{cool,dry,fan-only,heat,heat-cool,off}")
func (component *ModeToBool) SetValue(arg string) {
	component.Value.Set(arg == component.TrueValue)
}
