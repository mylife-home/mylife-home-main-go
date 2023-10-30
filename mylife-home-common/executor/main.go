package executor

import (
	"context"
	"mylife-home-common/tools"
	"sync"
)

// TODO: merge with plugin implementation

var ctxCancel func()
var ctx context.Context
var mainLoopChan *tools.ChannelMerger[func()]
var mainChan chan func()
var wg sync.WaitGroup

// Start the executor on the current goroutine
//
// Optionaly enqueue callbacks on startup
func Run(callbacks ...func()) {
	setup()

	// Cannot enqueue them right now because we are not buffered before loop start
	go func() {
		for _, callback := range callbacks {
			Execute(callback)
		}
	}()

	wg.Add(1)
	loop()
}

// Start the executor in its own goroutine
func Start() {
	setup()

	wg.Add(1)
	go loop()
}

func setup() {
	ctx, ctxCancel = context.WithCancel(context.Background())
	mainChan = make(chan func())
	mainLoopChan = tools.MakeChannelMerger[func()](mainChan)
}

// Stop the executor, optionally waiting for it to actually terminate its loop
// Note: if Stop(true) is called from inside the loop, it will deadlock.
//
// Exit Run() if executed from inside
func Stop(wait bool) {
	ctxCancel()
	close(mainChan)

	if wait {
		wg.Wait() // wait loop to exit
	}
}

// Queue callback for execution
func Execute(callback func()) {
	mainChan <- callback
}

// Get a context that is cancelled on termination
func Context() context.Context {
	return ctx
}

// Create a child executor. The main executor cannot terminate while there are children
func CreateExecutor() Executor {
	return &executorImpl{
		channel: mainLoopChan.Create(),
	}
}

func loop() {
	defer wg.Done()

	bufferedOut, bufferedIn := tools.BufferedChannel[func()]()
	tools.PipeChannel(mainLoopChan.Out(), bufferedOut)

	for callback := range bufferedIn {
		callback()
	}
}
