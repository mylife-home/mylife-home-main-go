package executor

import (
	"context"
	"time"
)

type Ticker struct {
	cancel func()
}

func NewTicker(duration time.Duration, callback func()) *Ticker {
	exec := CreateExecutor()
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer exec.Terminate()

		for {
			select {
			case <-time.After(duration):
				exec.Execute(callback)
			case <-Context().Done():
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return &Ticker{cancel}
}

func (ticker *Ticker) Stop() {
	ticker.cancel()
}
