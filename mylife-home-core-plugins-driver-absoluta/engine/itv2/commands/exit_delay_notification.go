package commands

import "mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

type ExitDelayNotification struct {
	Data *serialization.RemainBytes // ?? java states it should contains partition number, but Absoluta send something else
}

func init() {
	registerCommand[ExitDelayNotification](560)
}
