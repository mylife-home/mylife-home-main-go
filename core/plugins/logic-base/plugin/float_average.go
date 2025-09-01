package plugin

import (
	"math"
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type FloatAverage struct {
	// @Config(description="Nombre d'entrées qui sont utilisées (entre 0 et 8)")
	UsedCount int64

	state []float64

	// @State()
	Value definitions.State[float64]
}

func (component *FloatAverage) Init(runtime definitions.Runtime) error {
	component.state = make([]float64, component.UsedCount)

	for index := range component.state {
		component.state[index] = math.NaN()
	}

	component.Value.Set(math.NaN())
	return nil
}

func (component *FloatAverage) Terminate() {
	// Noop
}

func (component *FloatAverage) set(index int64, arg float64) {
	if index >= component.UsedCount {
		return
	}

	component.state[index] = arg

	value := math.NaN()

	if len(component.state) > 0 {
		var sum float64 = 0

		for _, item := range component.state {
			sum += item
		}

		value = sum / float64(len(component.state))
	}

	component.Value.Set(value)
}

// @Action()
func (component *FloatAverage) Set0(arg float64) {
	component.set(0, arg)
}

// @Action()
func (component *FloatAverage) Set1(arg float64) {
	component.set(1, arg)
}

// @Action()
func (component *FloatAverage) Set2(arg float64) {
	component.set(2, arg)
}

// @Action()
func (component *FloatAverage) Set3(arg float64) {
	component.set(3, arg)
}

// @Action()
func (component *FloatAverage) Set4(arg float64) {
	component.set(4, arg)
}

// @Action()
func (component *FloatAverage) Set5(arg float64) {
	component.set(5, arg)
}

// @Action()
func (component *FloatAverage) Set6(arg float64) {
	component.set(6, arg)
}

// @Action()
func (component *FloatAverage) Set7(arg float64) {
	component.set(7, arg)
}
