package plugin

import (
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-driver-tahoma/engine"
)

// @Plugin(usage="actuator")
type Box struct {
  // @Config(description="Identifiant pour que les composants soient mises à jour à partir de cette box Somfy")
  BoxKey string

  // @Config()
  User string

  // @Config()
  Password string
  
  // @State()
  Online bool

  store *Store
  client *Client
}

func (component *Box) Init() error {
  this.store = getStore(config.boxKey);

  this.client = new Client(config);
  this.store.setClient(this.client);
  this.store.on('onlineChanged', this.onOnlineChanged);

	return nil
}

func (component *Box) Terminate() {
  this.store.unsetClient();
  this.store.off('onlineChanged', this.onOnlineChanged);
  this.client.destroy();

  releaseStore(this.store.boxKey);
}

func (component *Box) onOnlineChanged(value bool) {
  this.online = value;
}
