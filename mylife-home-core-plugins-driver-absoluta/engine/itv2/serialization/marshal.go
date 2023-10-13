package serialization

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
)

type CustomMarshal interface {
	Write(writer BinaryWriter) error
	Read(reader BinaryReader) error
}

func Marshal(data *bytes.Buffer, object any) error {
	writer := makeBinaryWriter(data)
	return MarshalItem(writer, object)
}

func MarshalItem(writer BinaryWriter, object any) error {
	switch value := object.(type) {

	case *bool:
		if *value {
			return writer.WriteU8(1)
		} else {
			return writer.WriteU8(0)
		}
	case bool:
		if value {
			return writer.WriteU8(1)
		} else {
			return writer.WriteU8(0)
		}
	case []bool:
		for _, item := range value {
			if item {
				return writer.WriteU8(1)
			} else {
				return writer.WriteU8(0)
			}
		}

	case *int8:
		return writer.WriteI8(*value)
	case int8:
		return writer.WriteI8(value)
	case []int8:
		for _, item := range value {
			if err := writer.WriteI8(item); err != nil {
				return err
			}
		}
		return nil
	case *uint8:
		return writer.WriteU8(*value)
	case uint8:
		return writer.WriteU8(value)
	case []uint8:
		return writer.Write(value)

	case *int16:
		return writer.WriteI16(*value)
	case int16:
		return writer.WriteI16(value)
	case []int16:
		for _, item := range value {
			if err := writer.WriteI16(item); err != nil {
				return err
			}
		}
		return nil
	case *uint16:
		return writer.WriteU16(*value)
	case uint16:
		return writer.WriteU16(value)
	case []uint16:
		for _, item := range value {
			if err := writer.WriteU16(item); err != nil {
				return err
			}
		}
		return nil

	case *int32:
		return writer.WriteI32(*value)
	case int32:
		return writer.WriteI32(value)
	case []int32:
		for _, item := range value {
			if err := writer.WriteI32(item); err != nil {
				return err
			}
		}
		return nil
	case *uint32:
		return writer.WriteU32(*value)
	case uint32:
		return writer.WriteU32(value)
	case []uint32:
		for _, item := range value {
			if err := writer.WriteU32(item); err != nil {
				return err
			}
		}
		return nil

	case *int64:
		return writer.WriteI64(*value)
	case int64:
		return writer.WriteI64(value)
	case []int64:
		for _, item := range value {
			if err := writer.WriteI64(item); err != nil {
				return err
			}
		}
		return nil
	case *uint64:
		return writer.WriteU64(*value)
	case uint64:
		return writer.WriteU64(value)
	case []uint64:
		for _, item := range value {
			if err := writer.WriteU64(item); err != nil {
				return err
			}
		}
		return nil

	case *float32:
		return writer.WriteU32(math.Float32bits(*value))
	case float32:
		return writer.WriteU32(math.Float32bits(value))
	case []float32:
		for _, item := range value {
			if err := writer.WriteU32(math.Float32bits(item)); err != nil {
				return err
			}
		}
		return nil

	case *float64:
		return writer.WriteU64(math.Float64bits(*value))
	case float64:
		return writer.WriteU64(math.Float64bits(value))
	case []float64:
		for _, item := range value {
			if err := writer.WriteU64(math.Float64bits(item)); err != nil {
				return err
			}
		}
		return nil
	}

	custom, ok := object.(CustomMarshal)
	if ok {
		return custom.Write(writer)
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(object))

	switch reflectValue.Kind() {
	case reflect.Struct:
		for index := 0; index < reflectValue.NumField(); index++ {
			if err := MarshalItem(writer, reflectValue.Field(index).Interface()); err != nil {
				return err
			}
		}

		return nil

	case reflect.Array:
		for index := 0; index < reflectValue.Len(); index++ {
			if err := MarshalItem(writer, reflectValue.Index(index).Interface()); err != nil {
				return err
			}
		}

		return nil
	}

	return fmt.Errorf("cannot write value of type %s", reflect.TypeOf(object))
}

func Unmarshal(data *bytes.Buffer, object any) error {
	reader := makeBinaryReader(data)
	return UnmarshalItem(reader, object)
}

func UnmarshalItem(reader BinaryReader, object any) error {
	switch object := object.(type) {
	case *bool:
		val, err := reader.ReadU8()
		if err != nil {
			return err
		}
		*object = val != 0
		return nil
	case []bool:
		for index := range object {
			val, err := reader.ReadU8()
			if err != nil {
				return err
			}

			object[index] = val != 0
		}

	case *int8:
		val, err := reader.ReadI8()
		if err != nil {
			return err
		}
		*object = val
		return nil
	case []int8:
		for index := range object {
			val, err := reader.ReadI8()
			if err != nil {
				return err
			}

			object[index] = val
		}
	case *uint8:
		val, err := reader.ReadU8()
		if err != nil {
			return err
		}
		*object = val
		return nil
	case []uint8:
		for index := range object {
			val, err := reader.ReadU8()
			if err != nil {
				return err
			}

			object[index] = val
		}

	case *int16:
		val, err := reader.ReadI16()
		if err != nil {
			return err
		}
		*object = val
		return nil
	case []int16:
		for index := range object {
			val, err := reader.ReadI16()
			if err != nil {
				return err
			}

			object[index] = val
		}
	case *uint16:
		val, err := reader.ReadU16()
		if err != nil {
			return err
		}
		*object = val
		return nil
	case []uint16:
		for index := range object {
			val, err := reader.ReadU16()
			if err != nil {
				return err
			}

			object[index] = val
		}

	case *int32:
		val, err := reader.ReadI32()
		if err != nil {
			return err
		}
		*object = val
		return nil
	case []int32:
		for index := range object {
			val, err := reader.ReadI32()
			if err != nil {
				return err
			}

			object[index] = val
		}
	case *uint32:
		val, err := reader.ReadU32()
		if err != nil {
			return err
		}
		*object = val
		return nil
	case []uint32:
		for index := range object {
			val, err := reader.ReadU32()
			if err != nil {
				return err
			}

			object[index] = val
		}

	case *int64:
		val, err := reader.ReadI64()
		if err != nil {
			return err
		}
		*object = val
		return nil
	case []int64:
		for index := range object {
			val, err := reader.ReadI64()
			if err != nil {
				return err
			}

			object[index] = val
		}
	case *uint64:
		val, err := reader.ReadU64()
		if err != nil {
			return err
		}
		*object = val
		return nil
	case []uint64:
		for index := range object {
			val, err := reader.ReadU64()
			if err != nil {
				return err
			}

			object[index] = val
		}

	case *float32:
		val, err := reader.ReadU32()
		if err != nil {
			return err
		}
		*object = math.Float32frombits(val)
		return nil
	case []float32:
		for index := range object {
			val, err := reader.ReadU32()
			if err != nil {
				return err
			}

			object[index] = math.Float32frombits(val)
		}

	case *float64:
		val, err := reader.ReadU64()
		if err != nil {
			return err
		}
		*object = math.Float64frombits(val)
		return nil
	case []float64:
		for index := range object {
			val, err := reader.ReadU64()
			if err != nil {
				return err
			}

			object[index] = math.Float64frombits(val)
		}
	}

	reflectValue := reflect.ValueOf(object)

	if reflectValue.Kind() == reflect.Pointer {
		reflectValue = reflectValue.Elem()
	}

	if reflectValue.CanConvert(customMarshalType) {
		// create instance if nil, and read with interface
		if reflectValue.IsNil() {
			reflectValue.Set(reflect.New(reflectValue.Type().Elem()))
		}

		custom := reflectValue.Interface().(CustomMarshal)
		return custom.Read(reader)
	}

	switch reflectValue.Kind() {
	case reflect.Struct:
		for index := 0; index < reflectValue.NumField(); index++ {
			if err := UnmarshalItem(reader, reflectValue.Field(index).Addr().Interface()); err != nil {
				return err
			}
		}

		return nil

	case reflect.Array:
		for index := 0; index < reflectValue.Len(); index++ {
			if err := UnmarshalItem(reader, reflectValue.Index(index).Addr().Interface()); err != nil {
				return err
			}
		}

		return nil

	}

	return fmt.Errorf("cannot read value of type %s", reflect.TypeOf(object))
}

var customMarshalType = getType[CustomMarshal]()

func getType[T any]() reflect.Type {
	var ptr *T = nil
	return reflect.TypeOf(ptr).Elem()
}
