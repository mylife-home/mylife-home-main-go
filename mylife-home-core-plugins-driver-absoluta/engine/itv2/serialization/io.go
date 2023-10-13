package serialization

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type BinaryWriter interface {
	Write(data []byte) error
	WriteU8(value uint8) error
	WriteI8(value int8) error
	WriteU16(value uint16) error
	WriteI16(value int16) error
	WriteU32(value uint32) error
	WriteI32(value int32) error
	WriteU64(value uint64) error
	WriteI64(value int64) error
}

type BinaryReader interface {
	Read(data []byte) error
	ReadAll() ([]byte, error)
	Remain() int
	ReadU8() (uint8, error)
	ReadI8() (int8, error)
	ReadU16() (uint16, error)
	ReadI16() (int16, error)
	ReadU32() (uint32, error)
	ReadI32() (int32, error)
	ReadU64() (uint64, error)
	ReadI64() (int64, error)
}

func makeBinaryWriter(buffer *bytes.Buffer) BinaryWriter {
	return &binaryWriter{
		buffer: buffer,
	}
}

func makeBinaryReader(buffer *bytes.Buffer) BinaryReader {
	return &binaryReader{
		buffer: buffer,
	}
}

type binaryWriter struct {
	buffer *bytes.Buffer
}

func (writer *binaryWriter) Write(data []byte) error {
	_, err := writer.buffer.Write(data)
	return err
}

func (writer *binaryWriter) WriteU8(value uint8) error {
	return writer.buffer.WriteByte(value)
}

func (writer *binaryWriter) WriteI8(value int8) error {
	return writer.WriteU8(uint8(value))
}

func (writer *binaryWriter) WriteU16(value uint16) error {
	data := binary.BigEndian.AppendUint16(nil, value)
	return writer.Write(data)
}

func (writer *binaryWriter) WriteI16(value int16) error {
	return writer.WriteU16(uint16(value))
}

func (writer *binaryWriter) WriteU32(value uint32) error {
	data := binary.BigEndian.AppendUint32(nil, value)
	return writer.Write(data)
}

func (writer *binaryWriter) WriteI32(value int32) error {
	return writer.WriteU32(uint32(value))
}

func (writer *binaryWriter) WriteU64(value uint64) error {
	data := binary.BigEndian.AppendUint64(nil, value)
	return writer.Write(data)
}

func (writer *binaryWriter) WriteI64(value int64) error {
	return writer.WriteU64(uint64(value))
}

type binaryReader struct {
	buffer *bytes.Buffer
}

func (reader *binaryReader) Read(data []byte) error {
	if reader.buffer.Len() < len(data) {
		return fmt.Errorf("not enough data (read asked %d, buffer content %d)", len(data), reader.buffer.Len())
	}

	_, err := reader.buffer.Read(data)
	return err
}

func (reader *binaryReader) ReadAll() ([]byte, error) {
	data := make([]byte, reader.buffer.Len())
	if err := reader.Read(data); err != nil {
		return nil, err
	}

	return data, nil
}

func (reader *binaryReader) Remain() int {
	return reader.buffer.Len()
}

func (reader *binaryReader) ReadU8() (uint8, error) {
	return reader.buffer.ReadByte()
}

func (reader *binaryReader) ReadI8() (int8, error) {
	uval, err := reader.ReadU8()
	if err != nil {
		return 0, err
	}

	return int8(uval), nil
}

func (reader *binaryReader) ReadU16() (uint16, error) {
	data := make([]byte, 2)
	if err := reader.Read(data); err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(data), nil
}

func (reader *binaryReader) ReadI16() (int16, error) {
	uval, err := reader.ReadU16()
	if err != nil {
		return 0, err
	}

	return int16(uval), nil
}

func (reader *binaryReader) ReadU32() (uint32, error) {
	data := make([]byte, 4)
	if err := reader.Read(data); err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(data), nil
}

func (reader *binaryReader) ReadI32() (int32, error) {
	uval, err := reader.ReadU32()
	if err != nil {
		return 0, err
	}

	return int32(uval), nil
}

func (reader *binaryReader) ReadU64() (uint64, error) {
	data := make([]byte, 8)
	if err := reader.Read(data); err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(data), nil
}

func (reader *binaryReader) ReadI64() (int64, error) {
	uval, err := reader.ReadU64()
	if err != nil {
		return 0, err
	}

	return int64(uval), nil
}
