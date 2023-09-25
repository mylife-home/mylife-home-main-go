package tools

type ReadonlyMap[K comparable, V any] interface {
	Iterate() ReadonlyMapIterator[K, V]
	Len() int
	Get(key K) (V, bool)
	Clone() map[K]V
}

type ReadonlyMapIterator[K any, V any] interface {
	Next() bool
	Get() (K, V)
}

type ReadonlySlice[T any] interface {
	Iterate() ReadonlySliceIterator[T]
	Len() int
	Get(index int) T
	Clone() []T
}

type ReadonlySliceIterator[T any] interface {
	Next() bool
	Get() T
}

func NewReadonlyMap[K comparable, V any](m map[K]V) ReadonlyMap[K, V] {
	return &readonlyMap[K, V]{m}
}

func NewReadonlySlice[T any](s []T) ReadonlySlice[T] {
	return &readonlySlice[T]{s}
}

type readonlyMap[K comparable, V any] struct {
	target map[K]V
}

func (romap *readonlyMap[K, V]) Iterate() ReadonlyMapIterator[K, V] {
	kvpSlice := make([]keyValuePair[K, V], 0, len(romap.target))

	for key, value := range romap.target {
		kvpSlice = append(kvpSlice, keyValuePair[K, V]{key, value})
	}

	return newReadonlyMapIterator(kvpSlice)
}

func (romap *readonlyMap[K, V]) Len() int {
	return len(romap.target)
}

func (romap *readonlyMap[K, V]) Get(key K) (V, bool) {
	v, ok := romap.target[key]
	return v, ok
}

func (romap *readonlyMap[K, V]) Clone() map[K]V {
	m := make(map[K]V)

	for k, v := range romap.target {
		m[k] = v
	}

	return m
}

type readonlySlice[T any] struct {
	target []T
}

func (roslice *readonlySlice[T]) Iterate() ReadonlySliceIterator[T] {
	return newReadonlySliceIterator(roslice.target)
}

func (roslice *readonlySlice[T]) Len() int {
	return len(roslice.target)
}

func (roslice *readonlySlice[T]) Get(index int) T {
	return roslice.target[index]
}

func (roslice *readonlySlice[T]) Clone() []T {
	s := make([]T, len(roslice.target))

	for i, v := range roslice.target {
		s[i] = v
	}

	return s
}

type keyValuePair[K any, V any] struct {
	key   K
	value V
}

type readonlyMapIterator[K any, V any] struct {
	impl ReadonlySliceIterator[keyValuePair[K, V]]
}

func (iterator *readonlyMapIterator[K, V]) Next() bool {
	return iterator.impl.Next()
}

func (iterator *readonlyMapIterator[K, V]) Get() (K, V) {
	kvp := iterator.impl.Get()
	return kvp.key, kvp.value
}

func newReadonlyMapIterator[K any, V any](s []keyValuePair[K, V]) ReadonlyMapIterator[K, V] {
	return &readonlyMapIterator[K, V]{
		impl: newReadonlySliceIterator(s),
	}
}

type readonlySliceIterator[T any] struct {
	target  []T
	current int
}

func (iterator *readonlySliceIterator[T]) Next() bool {
	iterator.current += 1
	return iterator.current < len(iterator.target)
}

func (iterator *readonlySliceIterator[T]) Get() T {
	return iterator.target[iterator.current]
}

func newReadonlySliceIterator[T any](s []T) ReadonlySliceIterator[T] {
	return &readonlySliceIterator[T]{
		target:  s,
		current: -1,
	}
}
