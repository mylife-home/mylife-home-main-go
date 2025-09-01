package plugin

import (
	"mylife-home-common/tools"
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

	state           engine.State
	stateUpdateChan chan engine.StateValue
}

func (component *PartitionStatus) Init(runtime definitions.Runtime) error {
	component.Value.Set(false)

	component.state = engine.GetState(component.Key)

	component.stateUpdateChan = make(chan engine.StateValue)
	tools.DispatchChannel(component.stateUpdateChan, component.stateChanged)
	component.state.Subscribe(component.stateUpdateChan, true)

	return nil
}

func (component *PartitionStatus) Terminate() {
	component.state.Unsubscribe(component.stateUpdateChan)
	close(component.stateUpdateChan)
}

func (component *PartitionStatus) stateChanged(state engine.StateValue) {
	value := state.GetPartitionStatus(component.Label, component.Status)
	component.Value.Set(value)
}
