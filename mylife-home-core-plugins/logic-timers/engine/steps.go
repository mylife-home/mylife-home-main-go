package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type step interface {
	Execute(run *runningData)
}

var _ step = (*waitStep)(nil)

type waitStep struct {
	delay time.Duration
}

func newWaitStep(arg string) (step, error) {
	endOfDigits := strings.IndexFunc(arg, func(r rune) bool {
		return r < '0' || r > '9'
	})

	digits := arg
	suffix := ""

	if endOfDigits != -1 {
		digits = arg[:endOfDigits]
		suffix = arg[endOfDigits:]
	}

	delay, err := strconv.Atoi(digits)
	if err != nil {
		return nil, fmt.Errorf("invalid wait: '%s': %w", arg, err)
	}

	if suffix != "" {
		mul, ok := timerSuffixes[suffix]
		if !ok {
			return nil, fmt.Errorf("invalid wait: '%s': %w", arg, err)
		}

		delay *= mul
	}

	s := &waitStep{
		delay: time.Millisecond * time.Duration(delay),
	}

	return s, nil
}

func (s *waitStep) Execute(run *runningData) {
	logger.Debugf("Execute WaitStep: sleep %s", s.delay)

	remain := s.delay

	for remain > 0 {
		var sleep time.Duration
		if remain < time.Second {
			sleep = remain
		} else {
			sleep = time.Second
		}

		remain = remain - sleep

		select {
		case <-time.After(sleep):

		case <-run.ctx.Done():
			logger.Debug("WaitStep interrupted")
			return
		}

		run.computeProgress()
	}

	logger.Debug("WaitStep done")
}

var _ step = (*setOutputStep[int])(nil)

type setOutputStep[Value any] struct {
	setOutput func(index int, value Value)
	index     int
	value     Value
}

func newSetOutputStep[Value any](setOutput func(index int, value Value), index int, value Value) step {
	return &setOutputStep[Value]{
		setOutput: setOutput,
		index:     index,
		value:     value,
	}
}

func (s *setOutputStep[Value]) Execute(run *runningData) {
	logger.Debugf("Execute SetOutputStep: set output #%d to '%+v'", s.index, s.value)

	s.setOutput(s.index, s.value)
}

var _ step = (*setAllOutputsStep[int])(nil)

type setAllOutputsStep[Value any] struct {
	setOutput func(index int, value Value)
	value     Value
}

func newSetAllOutputsStep[Value any](setOutput func(index int, value Value), value Value) step {
	return &setAllOutputsStep[Value]{
		setOutput: setOutput,
		value:     value,
	}
}

func (s *setAllOutputsStep[Value]) Execute(run *runningData) {
	logger.Debugf("Execute SetAllOutputStep: set all outputs to '%+v'", s.value)

	for index := 0; index < outputCount; index += 1 {
		s.setOutput(index, s.value)
	}
}
