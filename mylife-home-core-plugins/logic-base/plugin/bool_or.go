package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type BoolOr struct {
	// @Config(description="Nombre d'entrées qui sont utilisées (entre 0 et 8)")
	UsedCount int64

	state []bool

	// @State()
	Value definitions.State[bool]
}

func (component *BoolOr) Init() error {
	component.state = make([]bool, component.UsedCount)

	return nil
}

func (component *BoolOr) Terminate() {
	// Noop
}

func (component *BoolOr) set(index int64, arg bool) {
	if index >= component.UsedCount {
		return
	}

	component.state[index] = arg

	var value bool = false
	for _, item := range component.state {
		value = value || item
	}

	component.Value.Set(value)
}

// @Action()
func (component *BoolOr) Set0(arg bool) {
	component.set(0, arg)
}

// @Action()
func (component *BoolOr) Set1(arg bool) {
	component.set(1, arg)
}

// @Action()
func (component *BoolOr) Set2(arg bool) {
	component.set(2, arg)
}

// @Action()
func (component *BoolOr) Set3(arg bool) {
	component.set(3, arg)
}

// @Action()
func (component *BoolOr) Set4(arg bool) {
	component.set(4, arg)
}

// @Action()
func (component *BoolOr) Set5(arg bool) {
	component.set(5, arg)
}

// @Action()
func (component *BoolOr) Set6(arg bool) {
	component.set(6, arg)
}

// @Action()
func (component *BoolOr) Set7(arg bool) {
	component.set(7, arg)
}
