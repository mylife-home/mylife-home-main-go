package engine

import (
	"mylife-home-common/tools"

	"github.com/mylife-home/klf200-go"
)

type Client struct {
	client *klf200.Client
	online tools.SubjectValue[bool]
}

func MakeClient(address string, password string) *Client {
	client := &Client{
		client: klf200.MakeClient(address, password, klf200logger),
		online: tools.MakeSubjectValue(false),
	}

	client.client.RegisterStatusChange(client.statusChange)

	client.client.Start()

	return client
}

func (client *Client) Terminate() {
	client.client.Close()
}

func (client *Client) statusChange(cs klf200.ConnectionStatus) {
	switch cs {
	case klf200.ConnectionOpen:
		client.online.Update(true)
	case klf200.ConnectionClosed, klf200.ConnectionHandshaking:
		client.online.Update(false)
	}
}

func (client *Client) Online() tools.ObservableValue[bool] {
	return client.online
}
