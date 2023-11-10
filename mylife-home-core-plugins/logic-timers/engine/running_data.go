package engine

import (
	"context"
	"time"
)

type runningData struct {
	// definition
	startTime      time.Time
	totalTime      time.Duration
	updateProgress func(time.Duration)
	exit           context.Context
}

func newRunningData[Value any](program *Program[Value]) *runningData {
	return &runningData{
		startTime:      time.Now(),
		totalTime:      program.totalTime,
		updateProgress: program.updateProgress,
	}
}

func (run *runningData) computeProgress() {
	if run.totalTime == 0 {
		panic("computeProgress called but totalTime = 0")
	}

	progressTime := time.Since(run.startTime)
	run.updateProgress(progressTime)
}

func (run *runningData) interrupted() <-chan struct{} {
	return run.exit.Done()
}
