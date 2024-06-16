package engine

import (
	"github.com/mylife-home/klf200-go/commands"
)

type Device struct {
	index   int
	address uint
	name    string
	typ     commands.ActuatorType
}

func (dev *Device) Index() int {
	return dev.index
}

func (dev *Device) Address() uint {
	return dev.address
}

func (dev *Device) Name() string {
	return dev.name
}

func (dev *Device) Type() commands.ActuatorType {
	return dev.typ
}

type DeviceState struct {
	deviceIndex int

	currentPosition int // percent, 0 = closed, 100 = open

	// TODO: add execution data
}

func (state *DeviceState) DeviceIndex() int {
	return state.deviceIndex
}

func (state *DeviceState) CurrentPosition() int {
	return state.currentPosition
}
