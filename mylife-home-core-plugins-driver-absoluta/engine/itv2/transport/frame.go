package transport

import (
	"bytes"
	"fmt"
	"io"
)

const frameStart byte = 0x7E
const frameEnd byte = 0x7F
const frameEscape byte = 0x7D

func MakeFrameChannelNode() *ChannelNode {
	return MakeChannelNode(&FrameEncoder{}, &FrameDecoder{})
}

type FrameEncoder struct {
	next Processor
}

func (encoder *FrameEncoder) SetNext(next Processor) {
	encoder.next = next
}

func (encoder *FrameEncoder) Process(input *bytes.Buffer) error {
	output := &bytes.Buffer{}

	if err := output.WriteByte(frameStart); err != nil {
		return err
	}

	for {
		b, err := input.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		if b == frameStart || b == frameEnd || b == frameEscape {
			if err := output.WriteByte(frameEscape); err != nil {
				return err
			}
			if err := output.WriteByte(b - frameEscape); err != nil {
				return err
			}
		} else {
			if err := output.WriteByte(b); err != nil {
				return err
			}
		}
	}

	if err := output.WriteByte(frameEnd); err != nil {
		return err
	}

	return encoder.next.Process(output)
}

type FrameDecoder struct {
	input *bytes.Buffer
	frame *bytes.Buffer
	next  Processor
}

func (decoder *FrameDecoder) SetNext(next Processor) {
	decoder.next = next
}

func (decoder *FrameDecoder) Process(data *bytes.Buffer) error {

	if decoder.input == nil {
		decoder.input = data
	} else {
		if _, err := decoder.input.Write(data.Bytes()); err != nil {
			return err
		}
	}

	for decoder.input.Len() > 0 {
		b, err := decoder.input.ReadByte()
		if err != nil {
			return err
		}

		switch b {
		case frameStart:
			if decoder.frame != nil {
				return fmt.Errorf("unexpected frame start")
			}

			decoder.frame = &bytes.Buffer{}

		case frameEnd:
			if decoder.frame == nil {
				return fmt.Errorf("unexpected frame end")
			}

			frame := decoder.frame
			decoder.frame = nil
			if err := decoder.next.Process(frame); err != nil {
				return err
			}

		case frameEscape:
			if decoder.frame == nil {
				return fmt.Errorf("unexpected data")
			}

			if decoder.input.Len() > 0 {
				b, err := decoder.input.ReadByte()
				if err != nil {
					return err
				}

				b += frameEscape
				if err := decoder.frame.WriteByte(b); err != nil {
					return err
				}
			} else {
				if err := decoder.input.UnreadByte(); err != nil {
					return err
				}

				return nil
			}

		default:
			if decoder.frame == nil {
				return fmt.Errorf("unexpected data")
			}

			if err := decoder.frame.WriteByte(b); err != nil {
				return err
			}
		}
	}

	decoder.input = nil
	return nil
}
