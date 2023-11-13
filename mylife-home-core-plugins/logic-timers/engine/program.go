package engine

import (
	"context"
	"mylife-home-common/log"
	"mylife-home-common/tools"
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

	onProgress tools.Subject[*ProgressArg]
	onRunning  tools.SubjectValue[bool]
	onOutput   tools.Subject[*OutputArg[Value]]
}

func NewProgram[Value any](parseOutputValue func(value string) (Value, error), source string, canWait bool) *Program[Value] {
	program := &Program[Value]{
		onProgress: tools.MakeSubject[*ProgressArg](),
		onRunning:  tools.MakeSubjectValue[bool](false),
		onOutput:   tools.MakeSubject[*OutputArg[Value]](),
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

func (program *Program[Value]) TotalTime() time.Duration {
	return program.totalTime
}

func (program *Program[Value]) OnProgress() tools.Observable[*ProgressArg] {
	return program.onProgress
}

func (program *Program[Value]) OnRunning() tools.ObservableValue[bool] {
	return program.onRunning
}

func (program *Program[Value]) OnOutput() tools.Observable[*OutputArg[Value]] {
	return program.onOutput
}

// true if run to the end, false if interrupted
// Note: start  several runs concurrently may end up with Observables mismatches
func (program *Program[Value]) Run(exit context.Context) bool {
	program.onRunning.Update(true)
	defer program.onRunning.Update(false)
	defer program.updateProgress(0)

	run := newRunningData(program, exit)

	for _, step := range program.steps {
		step.Execute(run)

		if exit.Err() != nil {
			return false
		}
	}

	return true

}

func (program *Program[Value]) updateProgress(progressTime time.Duration) {
	if program.totalTime == 0 {
		// do not emit progress on sync programs
		return
	}

	program.onProgress.Notify(&ProgressArg{
		percent:      float64(progressTime.Milliseconds()) / float64(program.totalTime.Milliseconds()) * 100,
		progressTime: progressTime,
	})
}

// Can be runned from any goroutine
func (program *Program[Value]) setOutput(index int, value Value) {
	program.onOutput.Notify(&OutputArg[Value]{
		index: index,
		value: value,
	})
}
