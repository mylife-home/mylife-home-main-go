package tools

import (
	"sync"

	"github.com/gookit/goutil/errorx/panics"
)

type Observable[T any] interface {
	Subscribe(observer chan<- T)
	Unsubscribe(observer chan<- T)
}

type Subject[T any] interface {
	Observable[T]

	Notify(event T)
}

type ObservableValue[T any] interface {
	Subscribe(observer chan<- T, sendCurrent bool)
	Unsubscribe(observer chan<- T)
	Get() T
}

type SubjectValue[T any] interface {
	ObservableValue[T]

	Update(newValue T) bool
}

var _ Observable[int] = (*subject[int])(nil)
var _ Subject[int] = (*subject[int])(nil)

type subject[T any] struct {
	observers map[chan<- T]struct{}
	mux       sync.RWMutex
}

func MakeSubject[T any]() Subject[T] {
	return &subject[T]{
		observers: make(map[chan<- T]struct{}),
	}
}

// Run all observers in parallel but wait for all to release the lock.
//
// Channels must be unbuffered to have synchronized dispatch.
//
// Channels must be buffered to have non-blocking dispatch.
func (sub *subject[T]) Notify(event T) {
	sub.mux.RLock()
	defer sub.mux.RUnlock()

	wg := sync.WaitGroup{}

	for observer := range sub.observers {
		wg.Add(1)

		go func(observer chan<- T) {
			defer wg.Done()
			observer <- event
		}(observer)
	}

	wg.Wait()

}

func (sub *subject[T]) Subscribe(observer chan<- T) {
	panics.IsTrue(observer != nil, "Subscribe with nil channel")

	sub.mux.Lock()
	defer sub.mux.Unlock()

	sub.observers[observer] = struct{}{}
}

func (sub *subject[T]) Unsubscribe(observer chan<- T) {
	panics.IsTrue(observer != nil, "Unsubscribe with nil channel")

	sub.mux.Lock()
	defer sub.mux.Unlock()

	delete(sub.observers, observer)
}

var _ ObservableValue[int] = (*subjectValue[int])(nil)
var _ SubjectValue[int] = (*subjectValue[int])(nil)

type subjectValue[T comparable] struct {
	observers map[chan<- T]struct{}
	value     T
	obsMux    sync.RWMutex
	valMux    sync.RWMutex
}

func MakeSubjectValue[T comparable](initial T) SubjectValue[T] {
	return &subjectValue[T]{
		observers: make(map[chan<- T]struct{}),
		value:     initial,
	}
}

func (sub *subjectValue[T]) Get() T {
	sub.valMux.RLock()
	defer sub.valMux.RUnlock()

	return sub.value
}

// Run all observers in parallel but wait for all to release the lock.
//
// Channels must be unbuffered to have synchronized dispatch.
//
// Channels must be buffered to have non-blocking dispatch.
func (sub *subjectValue[T]) Update(newValue T) bool {
	sub.valMux.Lock()
	defer sub.valMux.Unlock()

	if sub.value == newValue {
		return false
	}

	sub.value = newValue

	sub.obsMux.RLock()
	defer sub.obsMux.RUnlock()

	wg := sync.WaitGroup{}

	for observer := range sub.observers {
		wg.Add(1)

		go func(observer chan<- T) {
			defer wg.Done()
			observer <- sub.value
		}(observer)
	}

	wg.Wait()

	return true
}

func (sub *subjectValue[T]) Subscribe(observer chan<- T, sendCurrent bool) {
	panics.IsTrue(observer != nil, "Subscribe with nil channel")

	if sendCurrent {
		sub.valMux.RLock()
		defer sub.valMux.RUnlock()
	}

	sub.obsMux.Lock()
	defer sub.obsMux.Unlock()

	sub.observers[observer] = struct{}{}

	if sendCurrent {
		observer <- sub.value
	}
}

func (sub *subjectValue[T]) Unsubscribe(observer chan<- T) {
	panics.IsTrue(observer != nil, "Unsubscribe with nil channel")

	sub.obsMux.Lock()
	defer sub.obsMux.Unlock()

	delete(sub.observers, observer)
}
