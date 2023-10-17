package commands

import (
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"
)

type TroubleDetailNotification struct {
	Troubles *serialization.RemainArray[Trouble]
}

type Trouble struct {
	DeviceModuleType   *serialization.VarBytes
	TroubleType        *serialization.VarBytes
	DeviceModuleNumber *serialization.VarBytes
	Status             uint8
}

func (trouble *Trouble) GetDeviceModuleType() uint64 {
	return trouble.DeviceModuleType.GetUint()
}

func (trouble *Trouble) GetTroubleType() uint64 {
	return trouble.TroubleType.GetUint()
}

func (trouble *Trouble) GetDeviceModuleNumber() uint64 {
	return trouble.DeviceModuleNumber.GetUint()
}

func init() {
	registerCommand[TroubleDetailNotification](2083)
}
