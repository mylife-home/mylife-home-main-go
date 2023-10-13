package serialization

import (
	"bytes"
	"encoding/hex"
	"strings"
)

type RemainBytes struct {
	data []byte
}

var _ CustomMarshal = (*RemainBytes)(nil)

func (value *RemainBytes) Write(writer BinaryWriter) error {
	return writer.Write(value.data)
}

func (value *RemainBytes) Read(reader BinaryReader) error {
	buff, err := reader.ReadAll()
	if err != nil {
		return err
	}

	value.data = buff
	return nil
}

func (value *RemainBytes) Set(data []byte) {
	value.data = data
}

func (value *RemainBytes) Get() []byte {
	return value.data
}

func (value *RemainBytes) String() string {
	builder := strings.Builder{}
	builder.WriteString("[")
	for i, b := range value.data {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(hex.EncodeToString([]byte{b}))
	}
	builder.WriteString("]")
	return builder.String()
}

func RemainBytesGetItems[T any](value *RemainBytes, itemSize int, count int, factory func(data []byte) T) []T {
	buff := bytes.NewReader(value.data)
	array := make([]T, count)

	for index := 0; index < count; index += 1 {
		data := make([]byte, itemSize)
		n, err := buff.Read(data)
		if err != nil {
			panic(err)
		} else if n != itemSize {
			panic("Wrong read size")
		}

		array[index] = factory(data)
	}

	return array

}
