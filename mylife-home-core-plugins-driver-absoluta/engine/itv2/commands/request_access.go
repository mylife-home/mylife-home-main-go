package commands

import "mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

var _ Command = (*RequestAccess)(nil)
var _ CommandWithAppSeq = (*RequestAccess)(nil)

type RequestAccess struct {
	AppSeq uint8

	Identifier *serialization.VarBytes
}

func init() {
	registerCommand[RequestAccess](1550)
}

func (cmd *RequestAccess) GetAppSeq() uint8 {
	return cmd.AppSeq
}

func (cmd *RequestAccess) SetAppSeq(value uint8) {
	cmd.AppSeq = value
}
