package engine

import (
	"context"
	"time"
)

type runningData struct {
	// definition
	steps          []step
	startTime      time.Time
	totalTime      time.Duration
	updateProgress func(time.Duration)
	setRunning     func(bool)

	// runtime
	ctx       context.Context
	interrupt func()
	exited    chan struct{}
	onEnd     func() // when not interrupted
}

func newRun[Value any](program *Program[Value]) *runningData {
	ctx, interrupt := context.WithCancel(context.Background())

	run := &runningData{
		steps:          program.steps,
		startTime:      time.Now(),
		totalTime:      program.totalTime,
		updateProgress: program.updateProgress,
		setRunning:     program.setRunning,
		ctx:            ctx,
		interrupt:      interrupt,
		exited:         make(chan struct{}),
	}

	run.onEnd = func() {
		program.executeEnd(run)
	}

	return run
}

func (run *runningData) Start() {
	go run.Execute()
}

func (run *runningData) Stop() {
	run.interrupt()
	<-run.exited
}

func (run *runningData) Execute() {
	ended := run.executeSteps()

	if ended {
		// Normal end (no interrupt)
		run.onEnd()
	}
}

func (run *runningData) executeSteps() bool {
	defer close(run.exited)

	run.setRunning(true)
	defer run.setRunning(false)
	defer run.updateProgress(0)

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
	run.updateProgress(progressTime)
}
