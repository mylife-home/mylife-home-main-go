package plugin

import (
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-absoluta/engine"
)

// @Plugin(description="Fournit la valeur d'un indicateur composant l'état d'une zone" usage="sensor" version="1.0.0")
type ZoneStatus struct {
	// @Config(description="Clé de partage avec la connexion")
	Key string

	// @Config(description="Label de la zone")
	Label string

	// @Config(description="Nom de l'indicateur de status. Valeurs possibles: open, tamper, fault, lowBattery, delinquency, alarm, alarm-in-memory, by-passed")
	Status string

	// @State(description="current value")
	Value definitions.State[bool]

	state    *engine.State
	cbhandle engine.StateCallbackHandle
}

func (component *ZoneStatus) Init() error {
	component.Value.Set(false)

	component.state = engine.GetState(component.Key)
	component.cbhandle = component.state.ObserveChange(component.stateChanged)

	return nil
}

func (component *ZoneStatus) Terminate() {
	component.state.UnobserveChange(component.cbhandle)
}

func (component *ZoneStatus) stateChanged() {
	value := component.state.GetZoneStatus(component.Label, component.Status)
	component.Value.Set(value)
}
