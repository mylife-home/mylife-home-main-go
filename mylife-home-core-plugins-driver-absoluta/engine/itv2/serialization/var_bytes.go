package serialization

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strings"

	"github.com/gookit/goutil/errorx/panics"
)

type VarBytes struct {
	data []byte
}

var _ CustomMarshal = (*VarBytes)(nil)

func (value *VarBytes) Write(writer BinaryWriter) error {
	size := len(value.data)
	panics.IsTrue(size < 256)

	if err := writer.WriteU8(uint8(size)); err != nil {
		return err
	}

	return writer.Write(value.data)
}

func (value *VarBytes) Read(reader BinaryReader) error {
	size, err := reader.ReadU8()
	if err != nil {
		return err
	}

	buff := make([]byte, size)
	if err := reader.Read(buff); err != nil {
		return err
	}

	value.data = buff
	return nil
}

func (value *VarBytes) Set(data []byte) {
	value.data = data
}

func (value *VarBytes) Get() []byte {
	return value.data
}

func (value *VarBytes) String() string {
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

func (value *VarBytes) SetNull() {
	value.data = make([]byte, 0)
}

func (value *VarBytes) IsNull() bool {
	return len(value.data) == 0
}

func (value *VarBytes) SetUint(val uint64) {
	buf := &bytes.Buffer{}
	if val < 1<<8 {
		binary.Write(buf, binary.BigEndian, uint8(val))
	} else if val < 1<<16 {
		binary.Write(buf, binary.BigEndian, uint16(val))
	} else if val < 1<<32 {
		binary.Write(buf, binary.BigEndian, uint32(val))
	} else {
		binary.Write(buf, binary.BigEndian, val)
	}

	value.data = buf.Bytes()
}

func (value *VarBytes) GetUint() uint64 {
	if len(value.data) == 0 {
		return 0
	}

	// pad left and read it as uint64 BE
	buf := make([]byte, 8)
	copy(buf[8-len(value.data):8], value.data)
	return binary.BigEndian.Uint64(buf)
}
