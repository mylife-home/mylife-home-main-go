package definitions

import (
	"context"
)

type Runtime interface {
	ComponentId() string

	// Execute the callback on the same goroutine that executes actions
	//
	// This provides the ability to have mono-routine execution of operations inside the component
	// Can Execute/Terminate from anywhere
	// Will block component termination until it has been terminated
	NewExecutor() Executor

	// Get a context that is cancelled on component termination
	Context() context.Context
}

type Executor interface {
	Execute(callback func())
	Terminate()
}

/*
Example implementations

	// Note: could also keep one executor open and run each 'SetImmediate' as Execute on it
	func SetImmediate(runtime Runtime, callback func()) {
		e := runtime.NewExecutor()
		e.Execute(callback)
		e.Terminate()
	}

	func SetTimeout(runtime Runtime, duration time.Duration, callback func()) {
		e := runtime.NewExecutor()

		go func() {
			defer e.Terminate()

			select {
			case <-time.After(duration):
				e.Execute(callback)
			case <-runtime.Context().Done():
			}
		}()
	}

	func SetInterval(runtime Runtime, duration time.Duration, callback func()) {
		e := runtime.NewExecutor()

		go func() {
			defer e.Terminate()

			for {
				select {
				case <-time.After(duration):
					e.Execute(callback)
				case <-runtime.Context().Done():
					return
				}
			}
		}()
	}
*/
