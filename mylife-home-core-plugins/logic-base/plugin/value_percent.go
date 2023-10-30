package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ValuePercent struct {
	// @Config(name="toggleThreshold" description="Valeur partir de laquelle toggle passe à OFF ou ON. Typiquement 1 (Note: peut être écrasé par l'action 'setToggleThreshold'")
	ConfigToggleThreshold int64

	// @Config(name="onValue" description="Valeur définie lorsqu'on passe à ON. Typiquement 100 (Note: peut être écrasé par l'action 'setOnValue'")
	ConfigOnValue int64

	// @Config(name="offValue" description="Valeur définie lorsqu'on passe à OFF. Typiquement 0 (Note: peut être écrasé par l'action 'setOffValue'")
	ConfigOffValue int64

	// @State(type="range[0;100]" description="Valeur partir de laquelle toggle passe à OFF ou ON. Typiquement 1")
	ToggleThreshold definitions.State[int64]

	// @State(type="range[0;100]" description="Valeur définie lorsqu'on passe à ON. Typiquement 100")
	OnValue definitions.State[int64]

	// @State(type="range[0;100]" description="Valeur définie lorsqu'on passe à OFF. Typiquement 0")
	OffValue definitions.State[int64]

	// @State(type="range[0;100]")
	Value definitions.State[int64]
}

func (component *ValuePercent) Init(runtime definitions.Runtime) error {
	component.ToggleThreshold.Set(component.ConfigToggleThreshold)
	component.OnValue.Set(component.ConfigOnValue)
	component.OffValue.Set(component.ConfigOffValue)

	component.Value.Set(component.OffValue.Get())

	return nil
}

func (component *ValuePercent) Terminate() {
	// Noop
}

// @Action(type="range[0;100]")
func (component *ValuePercent) SetValue(arg int64) {
	component.Value.Set(arg)
}

// @Action(type="range[-1;100]")
func (component *ValuePercent) SetPulse(arg int64) {
	if arg != -1 {
		component.Value.Set(arg)
	}
}

// @Action()
func (component *ValuePercent) On(arg bool) {
	if arg {
		component.Value.Set(component.OnValue.Get())
	}
}

// @Action()
func (component *ValuePercent) Off(arg bool) {
	if arg {
		component.Value.Set(component.OffValue.Get())
	}
}

// @Action()
func (component *ValuePercent) Toggle(arg bool) {
	if arg {
		if component.Value.Get() < component.ToggleThreshold.Get() {
			component.Value.Set(component.OnValue.Get())
		} else {
			component.Value.Set(component.OffValue.Get())
		}
	}
}

// @Action(type="range[0;100]")
func (component *ValuePercent) SetToggleThreshold(arg int64) {
	component.ToggleThreshold.Set(arg)
}

// @Action(type="range[0;100]")
func (component *ValuePercent) SetOnValue(arg int64) {
	// Update value in case it is already onValue
	needUpdate := component.Value.Get() == component.OnValue.Get()

	component.OnValue.Set(arg)

	if needUpdate {
		component.Value.Set(component.OnValue.Get())
	}
}

// @Action(type="range[0;100]")
func (component *ValuePercent) SetOffValue(arg int64) {
	// Update value in case it is already offValue
	needUpdate := component.Value.Get() == component.OffValue.Get()

	component.OffValue.Set(arg)

	if needUpdate {
		component.Value.Set(component.OffValue.Get())
	}
}
