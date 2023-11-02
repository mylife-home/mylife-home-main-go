package tools

import "sync"

type ChannelMerger[T any] struct {
	out chan T
	wg  sync.WaitGroup
}

// Merge several channels into one.
// The output channel is closed when all input channels are closed.
// It allows to add input channels during its lifetime.
func MakeChannelMerger[T any](initialChan <-chan T) *ChannelMerger[T] {
	m := &ChannelMerger[T]{
		out: make(chan T),
	}

	m.Add(initialChan)

	// After first chan else we will close immediately
	go func() {
		m.wg.Wait()
		close(m.out)
	}()

	return m
}

func (m *ChannelMerger[T]) Add(ch <-chan T) {
	m.wg.Add(1)

	go func() {
		for value := range ch {
			m.out <- value
		}
		m.wg.Done()
	}()
}

func (m *ChannelMerger[T]) Create() chan<- T {
	ch := make(chan T)
	m.Add(ch)
	return ch
}

func (m *ChannelMerger[T]) Out() <-chan T {
	return m.out
}

// Create an infinite-buffered channel
func BufferedChannel[T any]() (chan<- T, <-chan T) {
	in := make(chan T)
	out := make(chan T)

	go func() {
		inputClosed := false
		buffer := make([]T, 0)
		defer close(out)

		for {
			var outChan chan T
			var outValue T
			if len(buffer) > 0 {
				outChan = out
				outValue = buffer[0]
			}

			var inChan chan T
			if !inputClosed {
				inChan = in
			}

			if outChan == nil && inChan == nil {
				// Nothing to do anymore, exit
				return
			}

			select {
			case value, ok := <-inChan:
				if !ok {
					inputClosed = true
					break
				}

				buffer = append(buffer, value)
				// TODO warn on buffer size too big

			case outChan <- outValue:
				buffer = buffer[1:]
			}
		}
	}()

	return in, out
}

// Connect one channel to another, with a mapper function between
func MapChannel[TIn any, TOut any](in <-chan TIn, mapper func(in TIn) TOut) <-chan TOut {
	out := make(chan TOut)

	go func() {
		defer close(out)

		for vin := range in {
			out <- mapper(vin)
		}
	}()

	return out
}

// Connect one channel to another, and only transmit values that pass the filter
func FilterChannel[T any](in <-chan T, filter func(in T) bool) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for vin := range in {
			if filter(vin) {
				out <- vin
			}
		}
	}()

	return out
}

// Connect one channel to another
func PipeChannel[T any](in <-chan T, out chan<- T) {
	go func() {
		defer close(out)

		for vin := range in {
			out <- vin
		}
	}()
}

// Read channel and ignore output until it has been closed
func DrainChannel[T any](in <-chan T) {
	for range in {
	}
}
