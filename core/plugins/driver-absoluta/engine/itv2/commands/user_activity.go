package commands

import "mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

var _ Command = (*UserActivity)(nil)
var _ CommandWithAppSeq = (*UserActivity)(nil)

type UserActivity struct {
	AppSeq uint8

	PartitionNumber *serialization.VarBytes
	Type            uint8
	UserCode        [0]byte // ??
}

func init() {
	registerCommand[UserActivity](2322)
}

func (cmd *UserActivity) GetAppSeq() uint8 {
	return cmd.AppSeq
}

func (cmd *UserActivity) SetAppSeq(value uint8) {
	cmd.AppSeq = value
}
