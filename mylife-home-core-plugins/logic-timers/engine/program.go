package engine

import (
	"mylife-home-common/log"
	"mylife-home-common/tools"
	"sync"
	"time"
)

var logger = log.CreateLogger("mylife:home:core:plugins:logic-timers:engine")

var timerSuffixes = map[string]int{
	"ms": 1,
	"s":  1000,
	"m":  60 * 1000,
	"h":  60 * 60 * 1000,
}

const outputCount = 8

type OutputArg[Value any] struct {
	index int
	value Value
}

func (arg *OutputArg[Value]) Index() int {
	return arg.index
}

func (arg *OutputArg[Value]) Value() Value {
	return arg.value
}

type ProgressArg struct {
	percent      float64
	progressTime time.Duration
}

func (arg *ProgressArg) Percent() float64 {
	return arg.percent
}

func (arg *ProgressArg) ProgressTime() time.Duration {
	return arg.progressTime
}

type Program[Value any] struct {
	// readonly after setup
	steps     []step
	totalTime time.Duration

	// used by run goroutine only
	onProgress *tools.CallbackManager[*ProgressArg]
	onRunning  *tools.CallbackManager[bool]
	onOutput   *tools.CallbackManager[*OutputArg[Value]]

	// updated by Start/Interrupt/End with sync
	running *runningData
	mux     sync.Mutex
}

func NewProgram[Value any](parseOutputValue func(value string) (Value, error), source string, canWait bool) *Program[Value] {
	program := &Program[Value]{
		onProgress: tools.NewCallbackManager[*ProgressArg](),
		onRunning:  tools.NewCallbackManager[bool](),
		onOutput:   tools.NewCallbackManager[*OutputArg[Value]](),
	}

	parser := ProgramParser[Value]{
		onOutput:         program.onOutput,
		parseOutputValue: parseOutputValue,
		source:           source,
		canWait:          canWait,
	}

	steps, err := parser.parse()
	if err != nil {
		logger.WithError(err).Error("Invalid program. Will fallback to empty program.")
		steps = make([]step, 0)
	}

	program.steps = steps

	var totalWait time.Duration

	for _, step := range steps {
		if ws, ok := step.(*waitStep); ok {
			totalWait += ws.delay
		}
	}

	program.totalTime = totalWait

	return program
}

func (program *Program[Value]) Start() {
	program.mux.Lock()
	defer program.mux.Unlock()

	program.ensureStopped()

	run := newRun(program)

	go run.execute()
	<-run.started
	program.running = run
}

func (program *Program[Value]) Interrupt() {
	program.mux.Lock()
	defer program.mux.Unlock()

	program.ensureStopped()
}

func (program *Program[Value]) ensureStopped() {
	if program.running == nil {
		return
	}

	program.running.interrupt()
	<-program.running.exited
	program.running = nil
}

func (program *Program[Value]) Terminate() {
	program.Interrupt()
}

func (program *Program[Value]) Running() bool {
	program.mux.Lock()
	defer program.mux.Unlock()

	return program.running != nil
}

func (program *Program[Value]) TotalTime() time.Duration {
	return program.totalTime
}

func (program *Program[Value]) executeEnd(run *runningData) {
	// We are after exit, so we check that no interruption occured in between before reseting
	program.mux.Lock()
	defer program.mux.Unlock()

	if program.running == run {
		program.running = nil
	}
}
