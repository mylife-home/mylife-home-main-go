package plugin

import (
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-notifications/engine"
	"strings"
)

// @Plugin(description="canal d'envoi de notifications par mail" usage="actuator")
type MailChannel struct {
	// @Config(description="Clé du canal")
	Key string

	// @Config(description="Serveur SMTP")
	SmtpServer string

	// @Config(description="Nom de connexion au serveur")
	User string

	// @Config(description="Mot de passe de connexion au serveur")
	Pass string

	// @Config(description="Liste de mails destinataires séparés par ';'")
	To string
}

func (channel *MailChannel) Init(runtime definitions.Runtime) error {
	engine.Register(channel.Key, engine.NewMailChannel(channel.SmtpServer, channel.User, channel.Pass, strings.Split(channel.To, ";")))

	return nil
}

func (channel *MailChannel) Terminate() {
	engine.Unregister(channel.Key)
}
