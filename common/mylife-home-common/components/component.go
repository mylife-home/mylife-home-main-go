package components

import (
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
)

type Component interface {
	Id() string
	Plugin() *metadata.Plugin

	StateItem(name string) tools.ObservableValue[any]
	Action(name string) chan<- any
}
