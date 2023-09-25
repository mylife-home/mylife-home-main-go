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
	client              *client
	tracking            bool
	onlineInstances     map[string]struct{}
	onlineInstancesSync sync.RWMutex
	onInstanceChange    *tools.CallbackManager[*InstancePresenceChange]
}

func newPresence(client *client, presenceTracking bool) *Presence {
	presence := &Presence{
		client:           client,
		tracking:         presenceTracking,
		onlineInstances:  make(map[string]struct{}),
		onInstanceChange: tools.NewCallbackManager[*InstancePresenceChange](),
	}

	if presence.tracking {
		presence.client.OnOnlineChanged().Register(presence.onOnlineChange)
		presence.client.OnMessage().Register(presence.onMessage)

		fireAndForget(func() error {
			return presence.client.Subscribe("+/online")
		})
	}

	return presence
}

func (presence *Presence) Tracking() bool {
	return presence.tracking
}

func (presence *Presence) OnInstanceChange() tools.CallbackRegistration[*InstancePresenceChange] {
	return presence.onInstanceChange
}

func (presence *Presence) IsOnline(instanceName string) bool {
	presence.onlineInstancesSync.RLock()
	defer presence.onlineInstancesSync.RUnlock()

	_, exists := presence.onlineInstances[instanceName]
	return exists
}

func (presence *Presence) GetOnlines() []string {
	presence.onlineInstancesSync.RLock()
	defer presence.onlineInstancesSync.RUnlock()

	return maps.Keys(presence.onlineInstances)
}

func (presence *Presence) onOnlineChange(online bool) {
	if !online {
		// No online instances anymore
		// Note: not thread safe, but we are called from mqtt routine, and offline right now
		for _, instanceName := range presence.GetOnlines() {
			presence.instanceChange(instanceName, false)
		}
	}
}

func (presence *Presence) onMessage(m *message) {
	if m.domain != presenceDomain {
		return
	}

	// if payload is empty, then this is a retain message deletion indicating that instance is offline
	online := len(m.Payload()) > 0 && Encoding.ReadBool(m.Payload())

	if m.InstanceName() == presence.client.InstanceName() {
		return
	}

	presence.instanceChange(m.InstanceName(), online)
}

func (presence *Presence) instanceChange(instanceName string, online bool) {
	presence.onlineInstancesSync.Lock()
	defer presence.onlineInstancesSync.Unlock()

	_, exists := presence.onlineInstances[instanceName]
	if online == exists {
		return
	}

	if online {
		presence.onlineInstances[instanceName] = struct{}{}
	} else {
		delete(presence.onlineInstances, instanceName)
	}

	presence.onInstanceChange.Execute(&InstancePresenceChange{instanceName, online})
}
