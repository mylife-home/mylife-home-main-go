package commands

import "mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

var _ Command = (*SectionRead)(nil)
var _ CommandWithAppSeq = (*SectionRead)(nil)

type SectionRead struct {
	AppSeq uint8
	Flags  *serialization.VarBytes // Empty for now
	// ModuleNumber
	MainSectionNumber uint16
	// SubSectionNumbers
	// Index
	// Count
}

func init() {
	registerCommand[SectionRead](1825)
}

func (cmd *SectionRead) GetAppSeq() uint8 {
	return cmd.AppSeq
}

func (cmd *SectionRead) SetAppSeq(value uint8) {
	cmd.AppSeq = value
}
