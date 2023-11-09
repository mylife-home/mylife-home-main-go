package bus

import (
	"mylife-home-common/tools"
	"sync"

	"golang.org/x/exp/maps"
)

type InstancePresenceChange struct {
	instanceName string
	online       bool
}

func (change *InstancePresenceChange) InstanceName() string {
	return change.instanceName
}

func (change *InstancePresenceChange) Online() bool {
	return change.online
}

const presenceDomain = "online"

type Presence struct {
	client *client

	instances map[string]struct{}
	onChange  tools.Subject[*InstancePresenceChange]
	mux       sync.RWMutex

	onlineChan   chan bool
	presenceChan chan *message
}

func newPresence(client *client) *Presence {
	presence := &Presence{
		client: client,

		instances: make(map[string]struct{}),
		onChange:  tools.MakeSubject[*InstancePresenceChange](),

		onlineChan:   make(chan bool),
		presenceChan: make(chan *message),
	}

	go presence.worker()
	presence.client.Online().Subscribe(presence.onlineChan, false)

	return presence
}

func (presence *Presence) terminate() {
	presence.client.Online().Unsubscribe(presence.onlineChan)

	if presence.client.Online().Get() {
		if err := presence.client.Unsubscribe("+/online"); err != nil {
			logger.WithError(err).Error("Cannot unsubscribe to presence topic")
		}
	}

	close(presence.onlineChan)
	close(presence.presenceChan)
}

func (presence *Presence) OnChange() tools.Observable[*InstancePresenceChange] {
	return presence.onChange
}

func (presence *Presence) IsOnline(instanceName string) bool {
	presence.mux.RLock()
	defer presence.mux.RUnlock()

	_, exists := presence.instances[instanceName]
	return exists
}

func (presence *Presence) GetOnlines() []string {
	presence.mux.RLock()
	defer presence.mux.RUnlock()

	return maps.Keys(presence.instances)
}

func (presence *Presence) onMessage(m *message) {
	presence.presenceChan <- m
}

func (presence *Presence) worker() {
	onlineChan := presence.onlineChan
	presenceChan := presence.presenceChan

	for onlineChan != nil || presenceChan != nil {

		select {
		case online, ok := <-onlineChan:
			if !ok {
				onlineChan = nil // closing
				continue
			}

			if online {
				go func() {
					if err := presence.client.Subscribe("+/online", presence.onMessage); err != nil {
						logger.WithError(err).Error("Cannot subscribe to presence topic")
					}
				}()
			} else {
				presence.clearOnlines()
			}

		case m, ok := <-presenceChan:
			if !ok {
				presenceChan = nil // closing
				continue
			}

			if m.InstanceName() == presence.client.InstanceName() {
				// ignore messages on self
				continue
			}

			// if payload is empty, then this is a retain message deletion indicating that instance is offline
			online := len(m.Payload()) > 0 && Encoding.ReadBool(m.Payload())

			presence.instanceChange(m.InstanceName(), online)
		}

	}
}

func (presence *Presence) clearOnlines() {
	presence.mux.Lock()
	defer presence.mux.Unlock()

	for instanceName := range presence.instances {
		delete(presence.instances, instanceName)
		presence.onChange.Notify(&InstancePresenceChange{instanceName, false})
	}
}

func (presence *Presence) instanceChange(instanceName string, online bool) {
	presence.mux.Lock()
	defer presence.mux.Unlock()

	_, exists := presence.instances[instanceName]
	if online == exists {
		return
	}

	if online {
		presence.instances[instanceName] = struct{}{}
	} else {
		delete(presence.instances, instanceName)
	}

	presence.onChange.Notify(&InstancePresenceChange{instanceName, online})
}
