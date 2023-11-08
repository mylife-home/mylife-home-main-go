package plugins

import (
	"fmt"
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
)

type untypedState interface {
	Value() tools.ObservableValue[any]
}

var _ definitions.State[int64] = (*stateImpl[int64])(nil)
var _ untypedState = (*stateImpl[int64])(nil)

type stateImpl[T comparable] struct {
	value tools.SubjectValue[any]
}

func (state *stateImpl[T]) Get() T {
	return state.value.Get().(T)
}

func (state *stateImpl[T]) Set(value T) {
	state.value.Update(value)
}

func (state *stateImpl[T]) Value() tools.ObservableValue[any] {
	return state.value
}

func (state *stateImpl[T]) init() {
	var defaultValue T
	state.value = tools.MakeSubjectValue[any](defaultValue)
}

type privateState interface {
	untypedState
	init()
}

func makeStateImpl(typ metadata.Type) untypedState {
	var state privateState
	switch typ.(type) {
	case *metadata.RangeType:
		state = &stateImpl[int64]{}
	case *metadata.TextType:
		state = &stateImpl[string]{}
	case *metadata.FloatType:
		state = &stateImpl[float64]{}
	case *metadata.BoolType:
		state = &stateImpl[bool]{}
	case *metadata.EnumType:
		state = &stateImpl[string]{}
	case *metadata.ComplexType:
		state = &stateImpl[any]{}
	default:
		panic(fmt.Sprintf("Unexpected type '%s'", typ.String()))
	}

	state.init()

	return state
}
