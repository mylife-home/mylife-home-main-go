package plugin

import (
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-absoluta/engine"
)

// @Plugin(description="connexion vers centrale d'alarme Bentel Absoluta" usage="sensor")
type Connection struct {
	// @Config(description="Clé de partage avec les autres composants")
	Key string

	// @Config(description="Adresse IP ou nom DNS de la centrale")
	ServerAddress string

	// @Config(description="Code pin de connexion")
	Pin string

	// @State(description="Indique si la connexion à la centrale est établie")
	Connected definitions.State[bool]

	service *engine.Service
}

func (component *Connection) Init(runtime definitions.Runtime) error {
	component.Connected.Set(false)

	state := engine.GetState(component.Key)
	component.service = engine.NewService(component.ServerAddress, component.Pin, state, component.connectedChanged)

	return nil
}

func (component *Connection) Terminate() {
	component.service.Terminate()
}

func (component *Connection) connectedChanged(newValue bool) {
	component.Connected.Set(newValue)
}
