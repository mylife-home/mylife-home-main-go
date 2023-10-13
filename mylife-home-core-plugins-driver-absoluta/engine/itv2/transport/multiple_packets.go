package transport

import (
	"bytes"

	"golang.org/x/exp/slices"
)

const multiplePacketsCommand uint16 = 1571

func MakeMultiplePacketChannelNode() *ChannelNode {
	return MakeChannelNode(nil, &MultiplePacketsDecoder{})
}

type MultiplePacketsDecoder struct {
	next Processor
}

func (decoder *MultiplePacketsDecoder) SetNext(next Processor) {
	decoder.next = next
}

func (decoder *MultiplePacketsDecoder) Process(input *bytes.Buffer) error {
	// Clone the command number to see if we can use it
	// TODO: not efficient
	clonedInput := bytes.NewBuffer(slices.Clone(input.Bytes()))
	cmd, err := ReadUint16BE(clonedInput)
	if err != nil {
		return err
	}

	if cmd != multiplePacketsCommand {
		return decoder.next.Process(input)
	}

	for clonedInput.Len() > 0 {
		size, err := DecodeSize(clonedInput)
		if err != nil {
			return err
		}

		data := make([]byte, size)

		if err := ReadBytes(clonedInput, data); err != nil {
			return err
		}

		if err := decoder.next.Process(bytes.NewBuffer(data)); err != nil {
			return err
		}
	}

	return nil
}
