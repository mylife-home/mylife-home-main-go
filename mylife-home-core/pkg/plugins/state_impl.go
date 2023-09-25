package plugins

import (
	"fmt"
	"mylife-home-common/components/metadata"
	"sync"
)

type untypedState interface {
	UntypedGet() any
	SetOnChange(value func(any))
}

type stateImpl[T comparable] struct {
	mutex    sync.Mutex
	value    T
	onChange func(any)
}

// State impl

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
		state.onChange(value)
	}
}

// untypedState impl

func (state *stateImpl[T]) UntypedGet() any {
	state.mutex.Lock()
	defer state.mutex.Unlock()

	return state.value
}

func (state *stateImpl[T]) SetOnChange(value func(any)) {
	if value == nil {
		value = func(value any) {}
	}

	state.onChange = value
}

// ---

func makeStateImpl(typ metadata.Type) untypedState {
	var state untypedState
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

	// setup noop
	state.SetOnChange(nil)

	return state
}
