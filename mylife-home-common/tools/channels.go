package tools

import "sync"

type ChannelMerger[T any] struct {
	out chan T
	wg  sync.WaitGroup
}

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

func (m *ChannelMerger[T]) Out() chan<- T {
	return m.out
}

func BufferedChannel[T any]() (<-chan T, chan<- T) {
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

func MapChannel[TIn any, TOut any](in <-chan TIn, out chan<- TOut, mapper func(in TIn) TOut) {
	go func() {
		for vin := range in {
			out <- mapper(vin)
		}
	}()
}
