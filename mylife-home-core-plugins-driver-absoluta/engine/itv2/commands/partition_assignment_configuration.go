package commands

import "mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

type PartitionAssignmentConfiguration struct {
	Partitions *serialization.VarBytes
}

func (cmd *PartitionAssignmentConfiguration) GetAssignedPartitions() []int {
	bits := serialization.NewBitMask(cmd.Partitions, 1, true)
	return bits.GetTrueIndexes()
}

func init() {
	registerCommand[PartitionAssignmentConfiguration](1906)
}
