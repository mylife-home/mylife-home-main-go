package plugin

import (
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-absoluta/engine"
)

// @Plugin(description="Fournit la valeur d'un indicateur composant l'état d'une partition" usage="sensor")
type PartitionStatus struct {
	// @Config(description="Clé de partage avec la connexion")
	Key string

	// @Config(description="Label de la partition")
	Label string

	// @Config(description="Nom de l'indicateur de status. Valeurs possibles: armed, stay, away, night, no-delay, alarm, troubles, alarm-in-memory, fire")
	Status string

	// @State(description="current value")
	Value definitions.State[bool]

	state    *engine.State
	cbhandle engine.StateCallbackHandle
}

func (component *PartitionStatus) Init() error {
	component.Value.Set(false)

	component.state = engine.GetState(component.Key)
	component.cbhandle = component.state.ObserveChange(component.stateChanged)

	return nil
}

func (component *PartitionStatus) Terminate() {
	component.state.UnobserveChange(component.cbhandle)
}

func (component *PartitionStatus) stateChanged() {
	value := component.state.GetPartitionStatus(component.Label, component.Status)
	component.Value.Set(value)
}
