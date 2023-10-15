package engine

import (
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/commands"
	"sync"

	"github.com/apex/log"
	"golang.org/x/exp/slices"
)

type StateCallbackHandle int

type State struct {
	partitions []*PartitionState
	zones      []*ZoneState
	mux        sync.RWMutex

	partitionLabels map[string]*PartitionState
	zoneLabels      map[string]*ZoneState

	cbmux     sync.Mutex
	cbcounter StateCallbackHandle
	callbacks map[StateCallbackHandle]func()
}

type PartitionState struct {
	label  string
	status commands.PartitionStatusWord
}

type ZoneState struct {
	label  string
	status commands.ZoneStatusWord
}

func makeState() *State {
	return &State{
		partitionLabels: make(map[string]*PartitionState),
		zoneLabels:      make(map[string]*ZoneState),
		cbcounter:       0,
		callbacks:       make(map[StateCallbackHandle]func()),
	}
}

func (state *State) ObserveChange(callback func()) StateCallbackHandle {
	state.cbmux.Lock()
	defer state.cbmux.Unlock()

	state.cbcounter += 1
	handle := state.cbcounter

	state.callbacks[handle] = callback
	return handle
}

func (state *State) UnobserveChange(handle StateCallbackHandle) {
	state.cbmux.Lock()
	defer state.cbmux.Unlock()

	delete(state.callbacks, handle)
}

func (state *State) notify() {
	state.cbmux.Lock()
	defer state.cbmux.Unlock()

	for _, callback := range state.callbacks {
		go callback()
	}
}

func (state *State) updateCapabilities(partitionCount int, zoneCount int) {
	// Note: should never change after first set
	state.mux.Lock()
	defer state.mux.Unlock()

	change := false

	if len(state.partitions) != partitionCount {
		state.partitions = make([]*PartitionState, partitionCount)
		state.partitionLabels = make(map[string]*PartitionState)
		change = true
	}

	if len(state.zones) != zoneCount {
		state.zones = make([]*ZoneState, zoneCount)
		state.zoneLabels = make(map[string]*ZoneState)
		change = true
	}

	if change {
		state.notify()
	}
}

func (state *State) updateAssignedPartitions(partitions []int) {
	// Note: change should be rare
	state.mux.Lock()
	defer state.mux.Unlock()

	partitionSet := make(map[int]struct{})
	for _, index := range partitions {
		// indexes starts at 1
		partitionSet[index-1] = struct{}{}
	}

	change := false

	for index := range state.partitions {
		_, assigned := partitionSet[index]
		actualAssigned := state.partitions[index] != nil
		if actualAssigned == assigned {
			continue
		}

		if assigned {
			state.partitions[index] = &PartitionState{}
		} else {
			oldLabel := state.partitions[index].label
			if oldLabel != "" {
				delete(state.partitionLabels, oldLabel)
			}

			state.partitions[index] = nil
		}

		change = true
	}

	if change {
		state.notify()
	}
}

func (state *State) updateAssignedZones(zones []int) {
	// Note: change should be rare
	state.mux.Lock()
	defer state.mux.Unlock()
	zoneSet := make(map[int]struct{})
	for _, index := range zones {
		// indexes starts at 1
		zoneSet[index-1] = struct{}{}
	}

	change := false

	for index := range state.zones {
		_, assigned := zoneSet[index]
		actualAssigned := state.zones[index] != nil
		if actualAssigned == assigned {
			continue
		}

		if assigned {
			state.zones[index] = &ZoneState{}
		} else {
			oldLabel := state.zones[index].label
			if oldLabel != "" {
				delete(state.zoneLabels, oldLabel)
			}

			state.zones[index] = nil
		}

		change = true
	}

	if change {
		state.notify()
	}
}

func (state *State) updatePartitionLabel(index int, label string) {
	// Note: change should be rare
	state.mux.Lock()
	defer state.mux.Unlock()

	change := false

	// indexes starts at 1
	partition := state.partitions[index-1]
	if partition == nil {
		logger.Warnf("Got label for unassigned partition %d", index)
		return
	}

	if partition.label != label {
		oldLabel := partition.label
		if oldLabel != "" {
			delete(state.partitionLabels, oldLabel)
		}

		partition.label = label
		state.partitionLabels[label] = partition
		logger.Debugf("Partition %d has label %s", index-1, label)

		change = true
	}

	if change {
		state.notify()
	}
}

func (state *State) updateZoneLabel(index int, label string) {
	// Note: change should be rare
	state.mux.Lock()
	defer state.mux.Unlock()

	change := false

	// indexes starts at 1
	zone := state.zones[index-1]
	if zone == nil {
		logger.Warnf("Got label for unassigned zone %d", index)
		return
	}

	if zone.label != label {
		oldLabel := zone.label
		if oldLabel != "" {
			delete(state.zoneLabels, oldLabel)
		}

		zone.label = label
		state.zoneLabels[label] = zone
		logger.Debugf("Zone %d has label %s", index-1, label)

		change = true
	}

	if change {
		state.notify()
	}
}

func (state *State) updatePartitionStatuses(statuses []commands.PartitionStatusWord) {
	// Note: change often
	state.mux.Lock()
	defer state.mux.Unlock()

	change := false

	for index, status := range statuses {
		partition := state.partitions[index]
		if partition != nil && !slices.Equal(partition.status, status) {
			partition.status = status
			change = true
		}
	}

	if change {
		logger.Debugf("Got partition statuses %+v", statuses)
		state.notify()
	}
}

func (state *State) updateZoneStatuses(statuses []commands.ZoneStatusWord) {
	// Note: change often
	state.mux.Lock()
	defer state.mux.Unlock()

	change := false

	for index, status := range statuses {
		zone := state.zones[index]
		if zone != nil && !slices.Equal(zone.status, status) {
			zone.status = status
			change = true
		}
	}

	if change {
		logger.Debugf("Got zone statuses %+v", statuses)
		state.notify()
	}
}

func (state *State) GetPartitionStatus(label string, statusPart string) bool {
	state.mux.RLock()
	defer state.mux.RUnlock()

	partition, ok := state.partitionLabels[label]
	if !ok {
		return false
	}

	status := partition.status
	if status == nil {
		return false
	}

	switch statusPart {
	case "armed":
		return status.Armed()
	case "stay":
		return status.Stay()
	case "away":
		return status.Away()
	case "night":
		return status.Night()
	case "no-delay":
		return status.NoDelay()
	case "alarm":
		return status.Alarm()
	case "troubles":
		return status.Troubles()
	case "alarm-in-memory":
		return status.AlarmInMemory()
	case "fire":
		return status.Fire()
	}

	log.Warnf("Unknown partition status '%s'", statusPart)

	return false
}

func (state *State) GetZoneStatus(label string, statusPart string) bool {
	state.mux.RLock()
	defer state.mux.RUnlock()

	zone, ok := state.zoneLabels[label]
	if !ok {
		return false
	}

	status := zone.status
	if status == nil {
		return false
	}

	switch statusPart {
	case "open":
		return status.Open()
	case "tamper":
		return status.Tamper()
	case "fault":
		return status.Fault()
	case "low-battery":
		return status.LowBattery()
	case "delinquency":
		return status.Delinquency()
	case "alarm":
		return status.Alarm()
	case "alarm-in-memory":
		return status.AlarmInMemory()
	case "by-passed":
		return status.ByPassed()
	}

	log.Warnf("Unknown zone status '%s'", statusPart)

	return false
}
