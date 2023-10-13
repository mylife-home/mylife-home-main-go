package transport

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"mylife-home-common/log"
)

var logger = log.CreateLogger("mylife:home:core:plugins:absoluta:engine:itv2:transport")

func EncodeSize(size int, output *bytes.Buffer) error {
	if size < 127 {
		return output.WriteByte(byte(size))
	}

	usize := uint16(size)
	usize |= 0x8000

	return WriteUint16BE(output, usize)
}

func DecodeSize(input *bytes.Buffer) (int, error) {
	b1, err := input.ReadByte()
	if err != nil {
		return 0, err
	}

	if (b1 & 0x80) == 0 {
		return int(b1), nil
	}

	b2, err := input.ReadByte()
	if err != nil {
		return 0, err
	}

	return int((uint16(b1)&0x7F)<<8 | uint16(b2)), nil
}

func ReadBytes(input *bytes.Buffer, data []byte) error {
	n, err := input.Read(data)
	if err != nil {
		return err
	}
	if n < len(data) {
		return fmt.Errorf("wrong read len")
	}

	return nil
}

func ReadUint16BE(input *bytes.Buffer) (uint16, error) {
	data := make([]byte, 2)
	if err := ReadBytes(input, data); err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(data), nil
}

func UnreadUint16(input *bytes.Buffer) error {
	return UnreadBytes(input, 2)
}

func UnreadBytes(input *bytes.Buffer, size int) error {
	for index := 0; index < size; index += 1 {
		if err := input.UnreadByte(); err != nil {
			return err
		}
	}

	return nil
}

func WriteBytes(output *bytes.Buffer, b []byte) error {
	_, err := output.Write(b)
	return err
}

func WriteUint16BE(output *bytes.Buffer, value uint16) error {
	data := binary.BigEndian.AppendUint16(nil, value)
	return WriteBytes(output, data)
}
