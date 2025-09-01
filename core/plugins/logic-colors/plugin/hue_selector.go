package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type HueSelector struct {

	// @Config()
	Hue0 int64

	// @Config()
	Hue1 int64

	// @Config()
	Hue2 int64

	// @Config()
	Hue3 int64

	// @Config()
	Hue4 int64

	// @Config()
	Hue5 int64

	// @Config()
	Hue6 int64

	// @Config()
	Hue7 int64

	// @Config()
	Hue8 int64

	// @Config()
	Hue9 int64

	// @State(type="range[0;255]")
	Hue definitions.State[int64]

	// @State()
	White definitions.State[bool]
}

func (component *HueSelector) Init(runtime definitions.Runtime) error {
	component.ensureBounds(&component.Hue0)
	component.ensureBounds(&component.Hue1)
	component.ensureBounds(&component.Hue2)
	component.ensureBounds(&component.Hue3)
	component.ensureBounds(&component.Hue4)
	component.ensureBounds(&component.Hue5)
	component.ensureBounds(&component.Hue6)
	component.ensureBounds(&component.Hue7)
	component.ensureBounds(&component.Hue8)
	component.ensureBounds(&component.Hue9)

	component.Hue.Set(component.Hue0)

	return nil
}

func (component *HueSelector) ensureBounds(config *int64) {
	if *config < 0 {
		*config = 0
	}

	if *config > 255 {
		*config = 255
	}
}

func (component *HueSelector) Terminate() {
	// Noop
}

// @Action()
func (component *HueSelector) SetWhite(arg bool) {
	if arg {
		component.White.Set(true)
	}
}

// @Action()
func (component *HueSelector) Set0(arg bool) {
	if arg {
		component.White.Set(false)
		component.Hue.Set(component.Hue0)
	}
}

// @Action()
func (component *HueSelector) Set1(arg bool) {
	if arg {
		component.White.Set(false)
		component.Hue.Set(component.Hue1)
	}
}

// @Action()
func (component *HueSelector) Set2(arg bool) {
	if arg {
		component.White.Set(false)
		component.Hue.Set(component.Hue2)
	}
}

// @Action()
func (component *HueSelector) Set3(arg bool) {
	if arg {
		component.White.Set(false)
		component.Hue.Set(component.Hue3)
	}
}

// @Action()
func (component *HueSelector) Set4(arg bool) {
	if arg {
		component.White.Set(false)
		component.Hue.Set(component.Hue4)
	}
}

// @Action()
func (component *HueSelector) Set5(arg bool) {
	if arg {
		component.White.Set(false)
		component.Hue.Set(component.Hue5)
	}
}

// @Action()
func (component *HueSelector) Set6(arg bool) {
	if arg {
		component.White.Set(false)
		component.Hue.Set(component.Hue6)
	}
}

// @Action()
func (component *HueSelector) Set7(arg bool) {
	if arg {
		component.White.Set(false)
		component.Hue.Set(component.Hue7)
	}
}

// @Action()
func (component *HueSelector) Set8(arg bool) {
	if arg {
		component.White.Set(false)
		component.Hue.Set(component.Hue8)
	}
}

// @Action()
func (component *HueSelector) Set9(arg bool) {
	if arg {
		component.White.Set(false)
		component.Hue.Set(component.Hue9)
	}
}
