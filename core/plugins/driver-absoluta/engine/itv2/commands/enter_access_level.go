package commands

import "mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

type EnterAccessLevelType = uint8

const (
	InstallerAccessLevel         EnterAccessLevelType = 0
	EnhancedInstallerAccessLevel EnterAccessLevelType = 1
	UserAccessLevel              EnterAccessLevelType = 2
)

var _ Command = (*EnterAccessLevel)(nil)
var _ CommandWithAppSeq = (*EnterAccessLevel)(nil)

type EnterAccessLevel struct {
	AppSeq uint8

	PartitionNumber       *serialization.VarBytes
	Type                  uint8
	ProgrammingAccessCode *serialization.VarBytes // BCD
}

func init() {
	registerCommand[EnterAccessLevel](1024)
}

func (cmd *EnterAccessLevel) GetAppSeq() uint8 {
	return cmd.AppSeq
}

func (cmd *EnterAccessLevel) SetAppSeq(value uint8) {
	cmd.AppSeq = value
}
