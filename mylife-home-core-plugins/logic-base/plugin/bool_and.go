package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type BoolAnd struct {
	// @Config(description="Nombre d'entrées qui sont utilisées (entre 0 et 8)")
	UsedCount int64

	state []bool

	// @State()
	Value definitions.State[bool]
}

func (component *BoolAnd) Init() error {
	component.state = make([]bool, component.UsedCount)

	component.Value.Set(true)
	return nil
}

func (component *BoolAnd) Terminate() {
	// Noop
}

func (component *BoolAnd) set(index int64, arg bool) {
	if index >= component.UsedCount {
		return
	}

	component.state[index] = arg

	var value bool = true
	for _, item := range component.state {
		value = value && item
	}

	component.Value.Set(value)
}

// @Action()
func (component *BoolAnd) Set0(arg bool) {
	component.set(0, arg)
}

// @Action()
func (component *BoolAnd) Set1(arg bool) {
	component.set(1, arg)
}

// @Action()
func (component *BoolAnd) Set2(arg bool) {
	component.set(2, arg)
}

// @Action()
func (component *BoolAnd) Set3(arg bool) {
	component.set(3, arg)
}

// @Action()
func (component *BoolAnd) Set4(arg bool) {
	component.set(4, arg)
}

// @Action()
func (component *BoolAnd) Set5(arg bool) {
	component.set(5, arg)
}

// @Action()
func (component *BoolAnd) Set6(arg bool) {
	component.set(6, arg)
}

// @Action()
func (component *BoolAnd) Set7(arg bool) {
	component.set(7, arg)
}
