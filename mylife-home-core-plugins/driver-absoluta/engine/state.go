package engine

import (
	"maps"
	"mylife-home-common/tools"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/commands"
	"sync"

	"github.com/apex/log"
	"golang.org/x/exp/slices"
)

type StateValue interface {
	GetPartitionStatus(label string, statusPart string) bool
	GetZoneStatus(label string, statusPart string) bool
}

type State = tools.ObservableValue[StateValue]

type partitionState struct {
	label  string
	status commands.PartitionStatusWord
}

type zoneState struct {
	label  string
	status commands.ZoneStatusWord
}

type stateValue struct {
	partitions []*partitionState
	zones      []*zoneState

	partitionLabels map[string]*partitionState
	zoneLabels      map[string]*zoneState
}

var initialStateValue StateValue = &stateValue{
	partitionLabels: make(map[string]*partitionState),
	zoneLabels:      make(map[string]*zoneState),
}

func (state *stateValue) updateCapabilities(partitionCount int, zoneCount int) *stateValue {
	// Note: should never change after first set
	change := false
	spartitions := state.partitions
	szones := state.zones
	spartitionLabels := state.partitionLabels
	szoneLabels := state.zoneLabels

	if len(spartitions) != partitionCount {
		spartitions = make([]*partitionState, partitionCount)
		spartitionLabels = make(map[string]*partitionState)
		change = true
	}

	if len(szones) != zoneCount {
		szones = make([]*zoneState, zoneCount)
		szoneLabels = make(map[string]*zoneState)
		change = true
	}

	if !change {
		return state
	}

	return &stateValue{
		partitions:      spartitions,
		zones:           szones,
		partitionLabels: spartitionLabels,
		zoneLabels:      szoneLabels,
	}
}

func (state *stateValue) updateAssignedPartitions(partitions []int) *stateValue {
	// Note: change should be rare

	partitionSet := make(map[int]struct{})
	for _, index := range partitions {
		// indexes starts at 1
		partitionSet[index-1] = struct{}{}
	}

	change := false
	spartitions := state.partitions
	spartitionLabels := state.partitionLabels

	for index := range spartitions {
		_, assigned := partitionSet[index]
		actualAssigned := spartitions[index] != nil
		if actualAssigned == assigned {
			continue
		}

		spartitions = slices.Clone(spartitions)

		if assigned {
			spartitions[index] = &partitionState{}
		} else {
			oldLabel := spartitions[index].label
			if oldLabel != "" {
				spartitionLabels = maps.Clone(spartitionLabels)
				delete(spartitionLabels, oldLabel)
			}

			spartitions[index] = nil
		}

		change = true
	}

	if !change {
		return state
	}

	return &stateValue{
		partitions:      spartitions,
		zones:           state.zones,
		partitionLabels: spartitionLabels,
		zoneLabels:      state.zoneLabels,
	}
}

func (state *stateValue) updateAssignedZones(zones []int) *stateValue {
	// Note: change should be rare

	zoneSet := make(map[int]struct{})
	for _, index := range zones {
		// indexes starts at 1
		zoneSet[index-1] = struct{}{}
	}

	change := false
	szones := state.zones
	szoneLabels := state.zoneLabels

	for index := range szones {
		_, assigned := zoneSet[index]
		actualAssigned := szones[index] != nil
		if actualAssigned == assigned {
			continue
		}

		szones = slices.Clone(szones)

		if assigned {
			szones[index] = &zoneState{}
		} else {
			oldLabel := szones[index].label
			if oldLabel != "" {
				szoneLabels = maps.Clone(szoneLabels)
				delete(szoneLabels, oldLabel)
			}

			szones[index] = nil
		}

		change = true
	}

	if !change {
		return state
	}

	return &stateValue{
		partitions:      state.partitions,
		zones:           szones,
		partitionLabels: state.partitionLabels,
		zoneLabels:      szoneLabels,
	}
}

func (state *stateValue) updatePartitionLabel(index int, label string) *stateValue {
	// Note: change should be rare

	// indexes starts at 1
	partition := state.partitions[index-1]
	if partition == nil {
		logger.Warnf("Got label for unassigned partition %d", index)
		return state
	}

	if partition.label == label {
		return state
	}

	spartitions := slices.Clone(state.partitions)
	spartitionLabels := maps.Clone(state.partitionLabels)
	*partition = *partition
	spartitions[index-1] = partition

	oldLabel := partition.label
	if oldLabel != "" {
		delete(spartitionLabels, oldLabel)
	}

	partition.label = label
	spartitionLabels[label] = partition
	logger.Debugf("Partition %d has label %s", index-1, label)

	return &stateValue{
		partitions:      spartitions,
		zones:           state.zones,
		partitionLabels: spartitionLabels,
		zoneLabels:      state.zoneLabels,
	}
}

func (state *stateValue) updateZoneLabel(index int, label string) *stateValue {
	// Note: change should be rare

	// indexes starts at 1
	zone := state.zones[index-1]
	if zone == nil {
		logger.Warnf("Got label for unassigned zone %d", index)
		return state
	}

	if zone.label == label {
		return state
	}

	szones := slices.Clone(state.zones)
	szoneLabels := maps.Clone(state.zoneLabels)
	*zone = *zone
	szones[index-1] = zone

	oldLabel := zone.label
	if oldLabel != "" {
		delete(szoneLabels, oldLabel)
	}

	zone.label = label
	szoneLabels[label] = zone
	logger.Debugf("Zone %d has label %s", index-1, label)

	return &stateValue{
		partitions:      state.partitions,
		zones:           szones,
		partitionLabels: state.partitionLabels,
		zoneLabels:      szoneLabels,
	}
}

func (state *stateValue) updatePartitionStatuses(statuses []commands.PartitionStatusWord) *stateValue {
	// Note: change often

	var spartitions []*partitionState

	for index, status := range statuses {
		partition := state.partitions[index]
		if partition == nil || slices.Equal(partition.status, status) {
			continue
		}

		// clone slice on first change only
		if spartitions == nil {
			spartitions = slices.Clone(state.partitions)
		}

		*partition = *partition
		spartitions[index] = partition

		partition.status = status
	}

	if spartitions == nil {
		return state
	}

	logger.Debugf("Got partition statuses %+v", statuses)

	return &stateValue{
		partitions:      spartitions,
		zones:           state.zones,
		partitionLabels: state.partitionLabels,
		zoneLabels:      state.zoneLabels,
	}
}

func (state *stateValue) updateZoneStatuses(statuses []commands.ZoneStatusWord) *stateValue {
	// Note: change often

	var szones []*zoneState

	for index, status := range statuses {
		zone := state.zones[index]
		if zone == nil || slices.Equal(zone.status, status) {
			continue
		}

		// clone slice on first change only
		if szones == nil {
			szones = slices.Clone(state.zones)
		}

		*zone = *zone
		szones[index] = zone

		zone.status = status
	}

	if szones == nil {
		return state
	}

	logger.Debugf("Got zone statuses %+v", statuses)

	return &stateValue{
		partitions:      state.partitions,
		zones:           szones,
		partitionLabels: state.partitionLabels,
		zoneLabels:      state.zoneLabels,
	}
}

func (state *stateValue) GetPartitionStatus(label string, statusPart string) bool {
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

func (state *stateValue) GetZoneStatus(label string, statusPart string) bool {
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

type stateUpdater struct {
	state tools.SubjectValue[StateValue]
	mux   sync.Mutex
}

// Note: avoid creating 2 updaters for 1 state, it would cause mux mismatches
func makeStateUpdater(state State) *stateUpdater {
	return &stateUpdater{
		state: state.(tools.SubjectValue[StateValue]),
	}
}

func (updater *stateUpdater) update(callback func(*stateValue) *stateValue) {
	updater.mux.Lock()
	defer updater.mux.Unlock()

	value := updater.state.Get().(*stateValue)

	newValue := callback(value)
	if newValue == value {
		return
	}

	updater.state.Update(newValue)

}

func (updater *stateUpdater) UpdateCapabilities(partitionCount int, zoneCount int) {
	updater.update(func(value *stateValue) *stateValue {
		return value.updateCapabilities(partitionCount, zoneCount)
	})
}

func (updater *stateUpdater) UpdateAssignedPartitions(partitions []int) {
	updater.update(func(value *stateValue) *stateValue {
		return value.updateAssignedPartitions(partitions)
	})
}

func (updater *stateUpdater) UpdateAssignedZones(zones []int) {
	updater.update(func(value *stateValue) *stateValue {
		return value.updateAssignedZones(zones)
	})
}

func (updater *stateUpdater) UpdatePartitionLabel(index int, label string) {
	updater.update(func(value *stateValue) *stateValue {
		return value.updatePartitionLabel(index, label)
	})
}

func (updater *stateUpdater) UpdateZoneLabel(index int, label string) {
	updater.update(func(value *stateValue) *stateValue {
		return value.updateZoneLabel(index, label)
	})
}

func (updater *stateUpdater) UpdatePartitionStatuses(statuses []commands.PartitionStatusWord) {
	updater.update(func(value *stateValue) *stateValue {
		return value.updatePartitionStatuses(statuses)
	})
}

func (updater *stateUpdater) UpdateZoneStatuses(statuses []commands.ZoneStatusWord) {
	updater.update(func(value *stateValue) *stateValue {
		return value.updateZoneStatuses(statuses)
	})
}
