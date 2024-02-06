package plugin

import (
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-notifications/engine"
)

// @Plugin(description="Notification a émettre" usage="actuator")
type Notification struct {
	// @Config(description="Clé du canal sur lequel émettre la notification")
	ChannelKey string

	// @Config(description="Titre de la notification")
	Title string

	// @Config(description="Texte de la notification")
	Text string
}

func (notif *Notification) Init(runtime definitions.Runtime) error {
	return nil
}

func (notif *Notification) Terminate() {
}

// @Action(description="Déclenchement de l'émission de la notification")
func (notif *Notification) Trigger(arg bool) {
	if arg {
		engine.Send(notif.ChannelKey, notif.Title, notif.Text)
	}
}
