package engine

import (
	"fmt"
	"strings"
)

type ProgramParser[Value any] struct {
	setOutput        func(index int, value Value)
	parseOutputValue func(value string) (Value, error)
	source           string
	canWait          bool
}

func (parser *ProgramParser[Value]) parse() ([]step, error) {
	steps := make([]step, 0)
	stepsConfig := strings.FieldsFunc(parser.source, func(r rune) bool {
		return r == '|' || r == ' '
	})

	for _, stepConfig := range stepsConfig {
		parts := strings.Split(stepConfig, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid step: '%s'", stepConfig)
		}

		op := parts[0]
		arg := parts[1]

		step, err := parser.createStep(op, arg)
		if err != nil {
			return nil, fmt.Errorf("invalid step: '%s': %w", stepConfig, err)
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func (parser *ProgramParser[Value]) createStep(op string, arg string) (step, error) {
	for index := 0; index < 10; index += 1 {
		if op == fmt.Sprintf("o%d", index) {
			value, err := parser.parseOutputValue(arg)
			if err != nil {
				return nil, err
			}

			step := newSetOutputStep[Value](parser.setOutput, index, value)
			return step, nil
		}
	}

	if op == "o*" {
		value, err := parser.parseOutputValue(arg)
		if err != nil {
			return nil, err
		}

		step := newSetAllOutputsStep[Value](parser.setOutput, value)
		return step, nil
	}

	if op == "w" {
		if !parser.canWait {
			return nil, fmt.Errorf("wait not allowed")
		}

		return newWaitStep(arg)
	}

	return nil, fmt.Errorf("invalid step operation: '%s'", op)
}
