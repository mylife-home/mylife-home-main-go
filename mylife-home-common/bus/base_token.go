package bus

import (
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// from mqtt implementation, it is private

type baseToken struct {
	m        sync.RWMutex
	complete chan struct{}
	err      error
}

// Wait implements the Token Wait method.
func (b *baseToken) Wait() bool {
	<-b.complete
	return true
}

// WaitTimeout implements the Token WaitTimeout method.
func (b *baseToken) WaitTimeout(d time.Duration) bool {
	timer := time.NewTimer(d)
	select {
	case <-b.complete:
		if !timer.Stop() {
			<-timer.C
		}
		return true
	case <-timer.C:
	}

	return false
}

// Done implements the Token Done method.
func (b *baseToken) Done() <-chan struct{} {
	return b.complete
}

func (b *baseToken) flowComplete() {
	select {
	case <-b.complete:
	default:
		close(b.complete)
	}
}

func (b *baseToken) Error() error {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.err
}

func (b *baseToken) setError(e error) {
	b.m.Lock()
	b.err = e
	b.flowComplete()
	b.m.Unlock()
}

func newBaseToken() *baseToken {
	return &baseToken{complete: make(chan struct{})}
}

func newDoneToken() mqtt.Token {
	token := newBaseToken()
	token.flowComplete()
	return token
}
