package commands

import (
	"fmt"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"
	"reflect"
)

var _ Command = (*Request)(nil)
var _ CommandWithAppSeq = (*Request)(nil)
var _ serialization.CustomMarshal = (*Request)(nil)

type Request struct {
	AppSeq  uint8
	ReqCode uint16
	ReqData any
}

func init() {
	registerCommand[Request](2048)
}

func (cmd *Request) GetAppSeq() uint8 {
	return cmd.AppSeq
}

func (cmd *Request) SetAppSeq(value uint8) {
	cmd.AppSeq = value
}

func (req *Request) Write(writer serialization.BinaryWriter) error {

	if err := writer.WriteU8(req.AppSeq); err != nil {
		return err
	}

	if err := writer.WriteU16(req.ReqCode); err != nil {
		return err
	}

	if err := serialization.MarshalItem(writer, req.ReqData); err != nil {
		return err
	}

	return nil
}

func (req *Request) Read(reader serialization.BinaryReader) error {
	panic("Cannot read request")
}

func (req *Request) String() string {
	return fmt.Sprintf("&{AppSeq:%d ReqData: %s %+v}", req.AppSeq, reflect.TypeOf(req.ReqData), req.ReqData)
}
