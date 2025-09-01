package commands

import "mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

type SystemCapabilities struct {
	MaxZones      *serialization.VarBytes
	MaxUsers      *serialization.VarBytes
	MaxPartitions *serialization.VarBytes
	MaxFOBs       *serialization.VarBytes
	MaxProxTags   *serialization.VarBytes
	MaxOutputs    *serialization.VarBytes
}

func init() {
	registerCommand[SystemCapabilities](1555)
}
