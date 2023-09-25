package components

import (
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
)

type StateChange struct {
	name  string
	value any
}

func (change *StateChange) Name() string {
	return change.name
}

func (change *StateChange) Value() any {
	return change.value
}

type Component interface {
	OnStateChange() tools.CallbackRegistration[*StateChange]

	Id() string
	Plugin() *metadata.Plugin

	ExecuteAction(name string, value any)
	GetStateItem(name string) any
	GetState() tools.ReadonlyMap[string, any]
}
