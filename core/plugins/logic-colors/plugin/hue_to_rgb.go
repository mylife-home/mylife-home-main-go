package plugin

import (
	"math"
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type HueToRgb struct {

	// @State(type="range[0;100]")
	Red definitions.State[int64]

	// @State(type="range[0;100]")
	Green definitions.State[int64]

	// @State(type="range[0;100]")
	Blue definitions.State[int64]

	// @State()
	White definitions.State[bool]

	// @State(type="range[0;255]")
	Hue definitions.State[int64]

	// @State(type="range[0;100]")
	Brightness definitions.State[int64]
}

func (component *HueToRgb) Init(runtime definitions.Runtime) error {
	component.compute()

	return nil
}

func (component *HueToRgb) Terminate() {
	// Noop
}

// @Action()
func (component *HueToRgb) SetWhite(arg bool) {
	component.White.Set(arg)
	component.compute()
}

// @Action(type="range[0;255]")
func (component *HueToRgb) SetHue(arg int64) {
	component.Hue.Set(arg)
	component.compute()
}

// @Action(type="range[0;100]")
func (component *HueToRgb) SetBrightness(arg int64) {
	component.Brightness.Set(arg)
	component.compute()
}

func (component *HueToRgb) compute() {
	if component.White.Get() {
		brightness := component.Brightness.Get()
		component.Red.Set(brightness)
		component.Green.Set(brightness)
		component.Blue.Set(brightness)
		return
	}

	// blue = 240deg => 0
	hue := math.Mod(((float64(component.Hue.Get()) * 360 / 255) + 240), 360)
	brightness := float64(component.Brightness.Get()) / 100
	r, g, b := hsv2rgb(hue, 1, brightness)
	component.Red.Set(int64(r * 100))
	component.Green.Set(int64(g * 100))
	component.Blue.Set(int64(b * 100))
}

// http://en.wikipedia.org/wiki/HSL_color_space
func hsv2rgb(h float64, s float64, v float64) (float64, float64, float64) {
	h /= 60
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h, 2)-1))
	if h < 1 {
		return c, x, 0
	}
	if h < 2 {
		return x, c, 0
	}
	if h < 3 {
		return 0, c, x
	}
	if h < 4 {
		return 0, x, c
	}
	if h < 5 {
		return x, 0, c
	}
	if h < 6 {
		return c, 0, x
	}

	// ??
	return 0, 0, 0
}
