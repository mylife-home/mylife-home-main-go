package definitions

type State[T any] interface {
	Get() T
	Set(value T)
}
