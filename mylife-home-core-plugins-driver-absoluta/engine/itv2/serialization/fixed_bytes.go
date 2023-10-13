package serialization

import (
	"encoding/hex"
	"strings"
)

// TODO: does not work on command decode (need size on construction)

type FixedBytes struct {
	data []byte
}

var _ CustomMarshal = (*FixedBytes)(nil)

func (value *FixedBytes) Write(writer BinaryWriter) error {
	return writer.Write(value.data)
}

func (value *FixedBytes) Read(reader BinaryReader) error {
	return reader.Read(value.data)
}

func (value *FixedBytes) Set(data []byte) {
	value.data = data
}

func (value *FixedBytes) Get() []byte {
	return value.data
}

func (value *FixedBytes) String() string {
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
