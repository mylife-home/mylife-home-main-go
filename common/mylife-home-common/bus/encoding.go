package bus

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"mylife-home-common/components/metadata"
	"reflect"
)

type encodingImpl struct{}

var Encoding = encodingImpl{}

func (e *encodingImpl) ReadString(buffer []byte) string {
	return string(buffer)
}

func (e *encodingImpl) WriteString(value string) []byte {
	return []byte(value)
}

func (e *encodingImpl) ReadBool(buffer []byte) bool {
	return buffer[0] != 0
}

func (e *encodingImpl) WriteBool(value bool) []byte {
	var iVal byte
	if value {
		iVal = 1
	} else {
		iVal = 0
	}

	return []byte{iVal}
}

func (e *encodingImpl) ReadUInt8(buffer []byte) uint8 {
	var data uint8
	e.read(buffer, &data)
	return data
}

func (e *encodingImpl) WriteUInt8(value uint8) []byte {
	return e.write(value)
}

func (e *encodingImpl) ReadInt8(buffer []byte) int8 {
	var data int8
	e.read(buffer, &data)
	return data
}

func (e *encodingImpl) WriteInt8(value int8) []byte {
	return e.write(value)
}

func (e *encodingImpl) ReadUInt32(buffer []byte) uint32 {
	var data uint32
	e.read(buffer, &data)
	return data
}

func (e *encodingImpl) WriteUInt32(value uint32) []byte {
	return e.write(value)
}

func (e *encodingImpl) ReadInt32(buffer []byte) int32 {
	var data int32
	e.read(buffer, &data)
	return data
}

func (e *encodingImpl) WriteInt32(value int32) []byte {
	return e.write(value)
}

func (e *encodingImpl) ReadFloat(buffer []byte) float64 {
	// Protocol uses float32
	var data float32
	e.read(buffer, &data)
	return float64(data)
}

func (e *encodingImpl) WriteFloat(value float64) []byte {
	// Protocol uses float32
	return e.write(float32(value))
}

func (e *encodingImpl) ReadJson(buffer []byte) any {
	var value any

	e.ReadTypedJson(buffer, &value)

	return value
}

func (e *encodingImpl) ReadTypedJson(buffer []byte, value any) {
	if err := json.Unmarshal(buffer, value); err != nil {
		panic(err)
	}
}

func (e *encodingImpl) WriteJson(value any) []byte {
	buf, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return buf
}

func (e *encodingImpl) read(buffer []byte, data any) {
	if err := binary.Read(bytes.NewReader(buffer), binary.LittleEndian, data); err != nil {
		panic(fmt.Errorf("could not read buffer %+v (%d) into type %s: %w", buffer, len(buffer), reflect.TypeOf(data), err))
	}
}

func (e *encodingImpl) write(data any) []byte {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (e *encodingImpl) WriteValue(typ metadata.Type, value any) []byte {
	switch realType := typ.(type) {
	case *metadata.RangeType:
		return e.writeRange(realType, value.(int64))

	case *metadata.TextType:
		return e.WriteString(value.(string))

	case *metadata.FloatType:
		return e.WriteFloat(value.(float64))

	case *metadata.BoolType:
		return e.WriteBool(value.(bool))

	case *metadata.EnumType:
		return e.WriteString(value.(string))

	case *metadata.ComplexType:
		return e.WriteJson(value)
	}

	panic(fmt.Errorf("unsupported type %s", typ.String()))
}

func (e *encodingImpl) ReadValue(typ metadata.Type, raw []byte) any {
	switch realType := typ.(type) {
	case *metadata.RangeType:
		return e.readRange(realType, raw)

	case *metadata.TextType:
		return e.ReadString(raw)

	case *metadata.FloatType:
		return e.ReadFloat(raw)

	case *metadata.BoolType:
		return e.ReadBool(raw)

	case *metadata.EnumType:
		return e.ReadString(raw)

	case *metadata.ComplexType:
		return e.ReadJson(raw)
	}

	panic(fmt.Errorf("unsupported type %s", typ.String()))
}

const int8Min int64 = -128
const int8Max int64 = 127
const uint8Max int64 = 255
const int32Min int64 = -2147483648
const int32Max int64 = 2147483647
const uint32Max int64 = 4294967295

func (e *encodingImpl) writeRange(typ *metadata.RangeType, value int64) []byte {
	if typ.Min() >= 0 && typ.Max() <= uint8Max {
		return e.WriteUInt8(uint8(value))
	}

	if typ.Min() >= int8Min && typ.Max() <= int8Max {
		return e.WriteInt8(int8(value))
	}

	if typ.Min() >= 0 && typ.Max() <= uint32Max {
		return e.WriteUInt32(uint32(value))
	}

	if typ.Min() >= int32Min && typ.Max() <= int32Max {
		return e.WriteInt32(int32(value))
	}

	panic(fmt.Errorf("Cannot represent range type with min=%d and max=%d because bounds are too big", typ.Min(), typ.Max()))
}

func (e *encodingImpl) readRange(typ *metadata.RangeType, raw []byte) int64 {
	if typ.Min() >= 0 && typ.Max() <= uint8Max {
		return int64(e.ReadUInt8(raw))
	}

	if typ.Min() >= int8Min && typ.Max() <= int8Max {
		return int64(e.ReadInt8(raw))
	}

	if typ.Min() >= 0 && typ.Max() <= uint32Max {
		return int64(e.ReadUInt32(raw))
	}

	if typ.Min() >= int32Min && typ.Max() <= int32Max {
		return int64(e.ReadInt32(raw))
	}

	panic(fmt.Errorf("Cannot represent range type with min=%d and max=%d because bounds are too big", typ.Min(), typ.Max()))
}
