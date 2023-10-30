package plugins

import (
	"fmt"
	"mylife-home-common/components/metadata"
	"mylife-home-core-library/definitions"
	"sync"
)

type untypedState interface {
}

type StateChange struct {
	ComponentId string
	StateName   string
	Value       any
}

var _ definitions.State[int64] = (*stateImpl[int64])(nil)
var _ untypedState = (*stateImpl[int64])(nil)

type stateImpl[T comparable] struct {
	mutex sync.Mutex
	value T

	onEmit func(any)
}

func (state *stateImpl[T]) Get() T {
	state.mutex.Lock()
	defer state.mutex.Unlock()

	return state.value
}

func (state *stateImpl[T]) Set(value T) {
	state.mutex.Lock()
	defer state.mutex.Unlock()

	if state.value != value {
		state.value = value
		state.emit()
	}
}

func (state *stateImpl[T]) emit() {
	state.onEmit(state.value)
}

func (state *stateImpl[T]) init(onEmit func(value any)) {
	state.onEmit = onEmit

	// emit initial state
	state.emit()
}

type privateState interface {
	untypedState
	init(onEmit func(value any))
}

func makeStateImpl(typ metadata.Type, onEmit func(value any)) untypedState {
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

	state.init(onEmit)

	return state
}
