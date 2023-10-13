package commands

import "mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

type AccessLevelLeadInOut struct {
	PartitionNumber *serialization.VarBytes // toPositiveInteger
	Type            uint8
	User            *serialization.VarBytes // toPositiveInt
	Access          uint8
	Mode            uint8
	Date            uint32
}

func init() {
	registerCommand[AccessLevelLeadInOut](1026)
}
