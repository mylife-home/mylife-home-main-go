package commands

import "mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

// Not implemented in java
// Receive 2 messages on zone status change:
// First contains [0 7 1]
// Second contains [0 7 0]
type ZoneStatusChangeNotification struct {
	Data *serialization.RemainBytes
}

func init() {
	registerCommand[ZoneStatusChangeNotification](578)
}
