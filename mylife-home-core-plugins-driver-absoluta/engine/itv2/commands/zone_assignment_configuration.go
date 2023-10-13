package commands

import "mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

var _ RequestData = (*ZoneAssignmentConfigurationRequest)(nil)
var _ ResponseData = (*ZoneAssignmentConfiguration)(nil)

type ZoneAssignmentConfigurationRequest struct {
	PartitionNumber *serialization.VarBytes
}

func (req *ZoneAssignmentConfigurationRequest) RequestCode() uint16 {
	return 1904
}

type ZoneAssignmentConfiguration struct {
	Req                 ZoneAssignmentConfigurationRequest
	PartitionAssignment *serialization.RemainBytes
}

func (cmd *ZoneAssignmentConfiguration) GetRequest() RequestData {
	return &cmd.Req
}

func (cmd *ZoneAssignmentConfiguration) GetAssignedZones() []int {
	bits := serialization.NewBitMask(cmd.PartitionAssignment, 1, true)
	return bits.GetTrueIndexes()
}

func init() {
	registerCommand[ZoneAssignmentConfiguration](1904)
}
