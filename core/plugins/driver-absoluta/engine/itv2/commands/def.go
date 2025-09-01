package commands

import (
	"bytes"
	"fmt"
	"mylife-home-common/log"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"
	"reflect"
)

var logger = log.CreateLogger("mylife:home:core:plugins:absoluta:engine:itv2:commands")

type Command interface {
}

type CommandWithAppSeq interface {
	Command
	GetAppSeq() uint8
	SetAppSeq(value uint8)
}

type RequestData interface {
	RequestCode() uint16
}

type ResponseData interface {
	Command
	GetRequest() RequestData
}

type Unknown struct {
	Code uint16
	Data []byte
}

var commandsByCode = make(map[uint16]reflect.Type)
var commandsByType = make(map[reflect.Type]uint16)

func registerCommand[T Command](code uint16) {
	var ptr *T = nil
	typ := reflect.TypeOf(ptr).Elem()

	_, exists := commandsByCode[code]
	if exists {
		panic(fmt.Sprintf("Command with code %d already registered", code))
	}

	_, exists = commandsByType[typ]
	if exists {
		panic(fmt.Sprintf("Command with type %s already registered", typ))
	}

	commandsByCode[code] = typ
	commandsByType[typ] = code

}

func DecodeCommand(data *bytes.Buffer) (Command, error) {
	var code uint16
	if err := serialization.Unmarshal(data, &code); err != nil {
		return nil, err
	}

	typ, ok := commandsByCode[code]
	if !ok {
		cmd := &Unknown{
			Code: code,
			Data: data.Bytes(),
		}

		return cmd, nil
	}

	command := reflect.New(typ).Interface().(Command)

	if err := serialization.Unmarshal(data, command); err != nil {
		return nil, fmt.Errorf("cannot read command of type '%s': %w", reflect.TypeOf(command), err)
	}

	return command, nil
}

func EncodeCommand(command Command) (*bytes.Buffer, error) {
	code, err := GetCommandCode(command)
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}

	if err := serialization.Marshal(buffer, code); err != nil {
		return nil, err
	}

	if err := serialization.Marshal(buffer, command); err != nil {
		return nil, err
	}

	return buffer, nil
}

func GetCommandCode(command Command) (uint16, error) {
	typ := reflect.TypeOf(command)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	code, ok := commandsByType[typ]
	if !ok {
		return 0, fmt.Errorf("unknown command type '%s'", typ)
	}

	return code, nil
}
