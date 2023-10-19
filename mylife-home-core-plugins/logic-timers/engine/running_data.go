package engine

import (
	"context"
	"mylife-home-common/tools"
	"time"
)

type runningData struct {
	// definition
	steps      []step
	startTime  time.Time
	totalTime  time.Duration
	onProgress *tools.CallbackManager[*ProgressArg]
	onRunning  *tools.CallbackManager[bool]

	// runtime
	ctx       context.Context
	interrupt func()
	started   chan struct{}
	exited    chan struct{}
	onEnd     func() // not interrupted
}

func newRun[Value any](program *Program[Value]) *runningData {
	ctx, interrupt := context.WithCancel(context.Background())

	run := &runningData{
		steps:      program.steps,
		startTime:  time.Now(),
		totalTime:  program.totalTime,
		onProgress: program.onProgress,
		onRunning:  program.onRunning,
		ctx:        ctx,
		interrupt:  interrupt,
		started:    make(chan struct{}),
		exited:     make(chan struct{}),
	}

	run.onEnd = func() {
		program.executeEnd(run)
	}

	return run
}

func (run *runningData) execute() {
	ended := run.executeSteps()

	if ended {
		// Normal end (no interrupt)
		run.onEnd()
	}
}

func (run *runningData) executeSteps() bool {
	close(run.started)
	defer close(run.exited)

	run.onRunning.Execute(true)
	defer run.onRunning.Execute(false)
	defer run.onProgress.Execute(&ProgressArg{percent: 0, progressTime: 0})

	for _, step := range run.steps {
		step.Execute(run)

		if run.ctx.Err() != nil {
			return false
		}
	}

	return true
}

func (run *runningData) computeProgress() {
	if run.totalTime == 0 {
		panic("computeProgress called but totalTime = 0")
	}

	progressTime := time.Since(run.startTime)
	percent := float64(progressTime.Milliseconds()) / float64(run.totalTime.Milliseconds()) * 100
	run.onProgress.Execute(&ProgressArg{percent: percent, progressTime: progressTime})
}
