package engine

import (
	"mylife-home-common/log"
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
	"time"

	"github.com/gookit/goutil/errorx/panics"
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
	executor  definitions.Executor

	onProgress *tools.CallbackManager[*ProgressArg]
	onRunning  *tools.CallbackManager[bool]
	onOutput   *tools.CallbackManager[*OutputArg[Value]]
	running    *runningData
}

func NewProgram[Value any](executor definitions.Executor, parseOutputValue func(value string) (Value, error), source string, canWait bool) *Program[Value] {
	program := &Program[Value]{
		executor:   executor,
		onProgress: tools.NewCallbackManager[*ProgressArg](),
		onRunning:  tools.NewCallbackManager[bool](),
		onOutput:   tools.NewCallbackManager[*OutputArg[Value]](),
	}

	parser := ProgramParser[Value]{
		setOutput:        program.setOutput,
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

func (program *Program[Value]) RunSync() {
	panics.IsNil(program.running)
	run := newRun(program)
	run.Execute()
}

func (program *Program[Value]) Start() {
	program.ensureStopped()

	run := newRun(program)
	run.Start()
	program.running = run
}

func (program *Program[Value]) Interrupt() {
	program.ensureStopped()
}

func (program *Program[Value]) ensureStopped() {
	if program.running == nil {
		return
	}

	program.running.Stop()
	program.running = nil
}

func (program *Program[Value]) Terminate() {
	program.ensureStopped()
}

func (program *Program[Value]) Running() bool {
	return program.running != nil
}

func (program *Program[Value]) TotalTime() time.Duration {
	return program.totalTime
}

// Called from different goroutine
func (program *Program[Value]) executeEnd(run *runningData) {
	program.executor.Execute(func() {
		panics.IsTrue(program.running == nil || program.running == run)
		program.running = nil
	})
}

func (program *Program[Value]) OnProgress() tools.CallbackRegistration[*ProgressArg] {
	return program.onProgress
}

func (program *Program[Value]) OnRunning() tools.CallbackRegistration[bool] {
	return program.onRunning
}

func (program *Program[Value]) OnOutput() tools.CallbackRegistration[*OutputArg[Value]] {
	return program.onOutput
}

// Can be runned from any goroutine
func (program *Program[Value]) updateProgress(progressTime time.Duration) {
	arg := &ProgressArg{percent: 0, progressTime: progressTime}
	if program.totalTime != 0 && progressTime != 0 {
		arg.percent = float64(progressTime.Milliseconds()) / float64(program.totalTime.Milliseconds()) * 100
	}

	program.executor.Execute(func() {
		program.onProgress.Execute(arg)
	})
}

// Can be runned from any goroutine
func (program *Program[Value]) setRunning(value bool) {
	program.executor.Execute(func() {
		program.onRunning.Execute(value)
	})
}

// Can be runned from any goroutine
func (program *Program[Value]) setOutput(index int, value Value) {
	program.onOutput.Execute(&OutputArg[Value]{
		index: index,
		value: value,
	})
}
