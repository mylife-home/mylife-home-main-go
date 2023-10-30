package tools

import (
	"golang.org/x/exp/maps"
)

type RegistrationToken int

const InvalidRegistrationToken RegistrationToken = 0

type CallbackRegistration[TArg any] interface {
	Register(callback func(TArg)) RegistrationToken
	Unregister(token RegistrationToken)
}

type CallbackManager[TArg any] struct {
	callbacks map[RegistrationToken]func(TArg)
	nextToken RegistrationToken
}

func NewCallbackManager[TArg any]() *CallbackManager[TArg] {
	return &CallbackManager[TArg]{
		callbacks: make(map[RegistrationToken]func(TArg)),
		nextToken: 1,
	}
}

func (m *CallbackManager[TArg]) Execute(arg TArg) {
	for _, callback := range m.cloneCallbacks() {
		callback(arg)
	}
}

// If callbacks are registered/unregistered inside executing, deadlock may appear without clone
func (m *CallbackManager[TArg]) cloneCallbacks() []func(TArg) {
	return maps.Values(m.callbacks)
}

func (m *CallbackManager[TArg]) Register(callback func(TArg)) RegistrationToken {
	token := m.nextToken
	m.nextToken += 1

	m.callbacks[token] = callback

	return token
}

func (m *CallbackManager[TArg]) Unregister(token RegistrationToken) {
	delete(m.callbacks, token)
}
