package tools

import "sync"

type Executor interface {
	// Execute the executor in this routing
	Run()

	// Start the executor in its own goroutine
	Start()

	// Stop the executor, optionally waiting for it to actually terminate its loop
	//
	// Exit Run() if executed from inside
	Stop(wait bool)

	// Queue callback for execution
	Execute(callback func())
}

// Setup for main program loop
var MainLoop Executor

var _ Executor = (*executor)(nil)

type executor struct {
	wg  sync.WaitGroup
	in  <-chan func()
	out chan<- func()
}

func NewExecutor() Executor {
	in, out := BufferedChannel[func()]()
	return &executor{
		in:  in,
		out: out,
	}
}

func (exec *executor) Run() {
	exec.wg.Add(1)
	exec.loop()
}

func (exec *executor) Start() {
	exec.wg.Add(1)
	go exec.loop()
}

func (exec *executor) loop() {
	defer exec.wg.Done()

	for callback := range exec.in {
		callback()
	}
}

func (exec *executor) Stop(wait bool) {
	close(exec.out)

	if wait {
		exec.wg.Wait()
	}
}

func (exec *executor) Execute(callback func()) {
	exec.out <- callback
}
