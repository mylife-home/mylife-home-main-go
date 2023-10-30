package executor

// Child executor
type Executor interface {
	// Execute on this child.
	//
	// This will run in the same goroutine as the main executor, but this one has its own lifecycle.
	//
	// Until Terminate() is called, you can call Execute(), even if the main executor is closing.
	// (But you are supposed to close ASAP)
	Execute(callback func())

	// Terminate this executor.
	//
	// No Execute() can be called on this executor once Terminte()
	Terminate()
}

var _ Executor = (*executorImpl)(nil)

type executorImpl struct {
	channel chan<- func()
}

func (exec *executorImpl) Execute(callback func()) {
	exec.channel <- callback
}

func (exec *executorImpl) Terminate() {
	close(exec.channel)
}
