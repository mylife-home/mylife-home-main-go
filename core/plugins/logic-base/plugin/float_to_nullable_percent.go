package plugin

import (
	"math"
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type FloatToNullablePercent struct {
	// @Config(description="Valeur d'entrée pour laquelle la sortie sera à 0%")
	Min float64

	// @Config(description="Valeur d'entrée pour laquelle la sortie sera à 100%. Note: si min > max, alors la conversion est dégressive.\nPar exemple si min = 10 et max = 0, alors 10 => 0%, 7.5 => 25%, 5 => 50%, 2.5 => 75%, 0 => 100%.")
	Max float64

	// @State(type="range[-1;100]")
	Value definitions.State[int64]
}

func (component *FloatToNullablePercent) Init(runtime definitions.Runtime) error {
	component.Value.Set(-1)

	return nil
}

func (component *FloatToNullablePercent) Terminate() {
	// Noop
}

// @Action()
func (component *FloatToNullablePercent) SetValue(arg float64) {
	if math.IsNaN(arg) {
		component.Value.Set(-1)
		return
	}

	reverse := component.Min > component.Max

	if reverse {
		component.Value.Set(100 - component.compute(component.Max, component.Min, arg))
	} else {
		component.Value.Set(component.compute(component.Min, component.Max, arg))
	}
}

func (component *FloatToNullablePercent) compute(min float64, max float64, arg float64) int64 {
	if arg < min {
		arg = min
	}

	if arg > max {
		arg = max
	}

	delta := max - min
	return int64(math.Round((arg - min) * 100 / delta))
}
