package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type Percent struct {

	// @Config()
	Value0 int64

	// @Config()
	Value1 int64

	// @Config()
	Value2 int64

	// @Config()
	Value3 int64

	// @Config()
	Value4 int64

	// @Config()
	Value5 int64

	// @Config()
	Value6 int64

	// @Config()
	Value7 int64

	// @Config()
	Value8 int64

	// @Config()
	Value9 int64

	// @Config()
	Step int64

	// @State(type="range[-1;100]")
	Value definitions.State[int64]

	loopbackValue int64
}

func (component *Percent) Init() error {
	component.ensureBounds(&component.Value0)
	component.ensureBounds(&component.Value1)
	component.ensureBounds(&component.Value2)
	component.ensureBounds(&component.Value3)
	component.ensureBounds(&component.Value4)
	component.ensureBounds(&component.Value5)
	component.ensureBounds(&component.Value6)
	component.ensureBounds(&component.Value7)
	component.ensureBounds(&component.Value8)
	component.ensureBounds(&component.Value9)

	if component.Step < 1 {
		component.Step = 1
	}

	if component.Step > 100 {
		component.Step = 100
	}

	component.Value.Set(-1)

	return nil
}

func (component *Percent) ensureBounds(config *int64) {
	if *config < 0 || *config > 100 {
		*config = -1
	}
}

func (component *Percent) Terminate() {
	// Noop
}

// @Action(type="range[0;100]")
func (component *Percent) SetValue(arg int64) {
	component.loopbackValue = arg
}

// @Action()
func (component *Percent) Up(arg bool) {
	if arg {
		value := component.loopbackValue + component.Step
		if value > 100 {
			value = 100
		}

		component.changeValue(value)
	}
}

// @Action()
func (component *Percent) Down(arg bool) {
	if arg {
		value := component.loopbackValue - component.Step
		if value < 0 {
			value = 0
		}

		component.changeValue(value)
	}
}

// @Action()
func (component *Percent) Set0(arg bool) {
	if arg {
		component.changeValue(component.Value0)
	}
}

// @Action()
func (component *Percent) Set1(arg bool) {
	if arg {
		component.changeValue(component.Value1)
	}
}

// @Action()
func (component *Percent) Set2(arg bool) {
	if arg {
		component.changeValue(component.Value2)
	}
}

// @Action()
func (component *Percent) Set3(arg bool) {
	if arg {
		component.changeValue(component.Value3)
	}
}

// @Action()
func (component *Percent) Set4(arg bool) {
	if arg {
		component.changeValue(component.Value4)
	}
}

// @Action()
func (component *Percent) Set5(arg bool) {
	if arg {
		component.changeValue(component.Value5)
	}
}

// @Action()
func (component *Percent) Set6(arg bool) {
	if arg {
		component.changeValue(component.Value6)
	}
}

// @Action()
func (component *Percent) Set7(arg bool) {
	if arg {
		component.changeValue(component.Value7)
	}
}

// @Action()
func (component *Percent) Set8(arg bool) {
	if arg {
		component.changeValue(component.Value8)
	}
}

// @Action()
func (component *Percent) Set9(arg bool) {
	if arg {
		component.changeValue(component.Value9)
	}
}

func (component *Percent) changeValue(value int64) {
	component.Value.Set(value)
	component.Value.Set(-1)
}
