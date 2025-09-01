package plugin

import (
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-absoluta/engine"
)

// @Plugin(description="Fournit la valeur d'un indicateur composant l'état d'une zone" usage="sensor")
type ZoneStatus struct {
	// @Config(description="Clé de partage avec la connexion")
	Key string

	// @Config(description="Label de la zone")
	Label string

	// @Config(description="Nom de l'indicateur de status. Valeurs possibles: open, tamper, fault, lowBattery, delinquency, alarm, alarm-in-memory, by-passed")
	Status string

	// @State(description="current value")
	Value definitions.State[bool]

	state           engine.State
	stateUpdateChan chan engine.StateValue
}

func (component *ZoneStatus) Init(runtime definitions.Runtime) error {
	component.Value.Set(false)

	component.state = engine.GetState(component.Key)

	component.stateUpdateChan = make(chan engine.StateValue)
	tools.DispatchChannel(component.stateUpdateChan, component.stateChanged)
	component.state.Subscribe(component.stateUpdateChan, true)

	return nil
}

func (component *ZoneStatus) Terminate() {
	component.state.Unsubscribe(component.stateUpdateChan)
	close(component.stateUpdateChan)
}

func (component *ZoneStatus) stateChanged(state engine.StateValue) {
	value := state.GetZoneStatus(component.Label, component.Status)
	component.Value.Set(value)
}
