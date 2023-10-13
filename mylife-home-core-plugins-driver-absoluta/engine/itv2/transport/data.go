package transport

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/sigurn/crc16"
)

func MakeDataChannelNode() *ChannelNode {
	return MakeChannelNode(&DataEncoder{}, &DataDecoder{})
}

type DataEncoder struct {
	next Processor
}

func (encoder *DataEncoder) SetNext(next Processor) {
	encoder.next = next
}

var crcTable = crc16.MakeTable(crc16.CRC16_CCITT_FALSE)

func (encoder *DataEncoder) Process(input *bytes.Buffer) error {
	output := &bytes.Buffer{}

	size := input.Len() + 2
	if err := EncodeSize(size, output); err != nil {
		return err
	}

	if _, err := output.Write(input.Bytes()); err != nil {
		return err
	}

	crc := crc16.Checksum(output.Bytes(), crcTable)

	if _, err := output.Write(binary.BigEndian.AppendUint16(nil, crc)); err != nil {
		return err
	}

	return encoder.next.Process(output)
}

type DataDecoder struct {
	next Processor
}

func (decoder *DataDecoder) SetNext(next Processor) {
	decoder.next = next
}

func (decoder *DataDecoder) Process(input *bytes.Buffer) error {

	size16, err := DecodeSize(input)
	if err != nil {
		return err
	}

	size := int(size16)

	crcw := crc16.New(crcTable)

	// TODO: use raw
	sizeBuff := &bytes.Buffer{}
	EncodeSize(size, sizeBuff)

	if _, err := crcw.Write(sizeBuff.Bytes()); err != nil {
		return err
	}

	if input.Len() < size {
		return fmt.Errorf("message too short (msg len = %d, header len = %d)", input.Len(), size)
	}

	data := make([]byte, size-2)
	if err := ReadBytes(input, data); err != nil {
		return err
	}

	if _, err := crcw.Write(data); err != nil {
		return err
	}

	crcData := make([]byte, 2)

	if err := ReadBytes(input, crcData); err != nil {
		return err
	}

	crcExpected := crcw.Sum16()
	crcActual := binary.BigEndian.Uint16(crcData)

	if crcActual != crcExpected {
		return fmt.Errorf("CRC error (actual=%04X, expected=%04X)", crcActual, crcExpected)
	}

	return decoder.next.Process(bytes.NewBuffer(data))
}
