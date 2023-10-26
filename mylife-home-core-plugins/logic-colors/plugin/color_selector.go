package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ColorSelector struct {

	// @Config()
	Color0 int64

	// @Config()
	Color1 int64

	// @Config()
	Color2 int64

	// @Config()
	Color3 int64

	// @Config()
	Color4 int64

	// @Config()
	Color5 int64

	// @Config()
	Color6 int64

	// @Config()
	Color7 int64

	// @Config()
	Color8 int64

	// @Config()
	Color9 int64

	// @State(type="range[0;16777215]")
	Color definitions.State[int64]
}

func (component *ColorSelector) Init(runtime definitions.Runtime) error {
	component.ensureBounds(&component.Color0)
	component.ensureBounds(&component.Color1)
	component.ensureBounds(&component.Color2)
	component.ensureBounds(&component.Color3)
	component.ensureBounds(&component.Color4)
	component.ensureBounds(&component.Color5)
	component.ensureBounds(&component.Color6)
	component.ensureBounds(&component.Color7)
	component.ensureBounds(&component.Color8)
	component.ensureBounds(&component.Color9)

	component.Color.Set(component.Color0)

	return nil
}

func (component *ColorSelector) ensureBounds(config *int64) {
	if *config < 0 {
		*config = 0
	}

	if *config > 16777215 {
		*config = 16777215
	}
}

func (component *ColorSelector) Terminate() {
	// Noop
}

// @Action()
func (component *ColorSelector) Set0(arg bool) {
	if arg {
		component.Color.Set(component.Color0)
	}
}

// @Action()
func (component *ColorSelector) Set1(arg bool) {
	if arg {
		component.Color.Set(component.Color1)
	}
}

// @Action()
func (component *ColorSelector) Set2(arg bool) {
	if arg {
		component.Color.Set(component.Color2)
	}
}

// @Action()
func (component *ColorSelector) Set3(arg bool) {
	if arg {
		component.Color.Set(component.Color3)
	}
}

// @Action()
func (component *ColorSelector) Set4(arg bool) {
	if arg {
		component.Color.Set(component.Color4)
	}
}

// @Action()
func (component *ColorSelector) Set5(arg bool) {
	if arg {
		component.Color.Set(component.Color5)
	}
}

// @Action()
func (component *ColorSelector) Set6(arg bool) {
	if arg {
		component.Color.Set(component.Color6)
	}
}

// @Action()
func (component *ColorSelector) Set7(arg bool) {
	if arg {
		component.Color.Set(component.Color7)
	}
}

// @Action()
func (component *ColorSelector) Set8(arg bool) {
	if arg {
		component.Color.Set(component.Color8)
	}
}

// @Action()
func (component *ColorSelector) Set9(arg bool) {
	if arg {
		component.Color.Set(component.Color9)
	}
}
