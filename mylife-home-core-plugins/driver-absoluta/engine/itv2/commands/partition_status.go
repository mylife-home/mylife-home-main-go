package commands

import (
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"
	"strings"
)

const (
	// Note: this are indexes on status words
	PartitionStatusArmed         int = 0
	PartitionStatusStay          int = 1
	PartitionStatusAway          int = 2
	PartitionStatusNight         int = 3
	PartitionStatusNoDelay       int = 4
	PartitionStatusAlarm         int = 8
	PartitionStatusTroubles      int = 9
	PartitionStatusAlarmInMemory int = 12
	PartitionStatusFire          int = 17
)

var _ RequestData = (*PartitionStatusRequest)(nil)
var _ ResponseData = (*PartitionStatus)(nil)

type PartitionStatusRequest struct {
	Partitions *serialization.VarBytes
}

func (req *PartitionStatusRequest) RequestCode() uint16 {
	return 2066
}

type PartitionStatus struct {
	Req           PartitionStatusRequest // Bitmask
	BytesOfStatus uint8
	Statuses      *serialization.RemainBytes // TODO: DscArray<DscBitMask>, array size = number of partitions, bitmask size = BytesOfStatus
}

func (cmd *PartitionStatus) GetRequest() RequestData {
	return &cmd.Req
}

func (cmd *PartitionStatus) GetPartitions() []int {
	return serialization.NewBitMask(cmd.Req.Partitions, 0, true).GetTrueIndexes()
}

func (cmd *PartitionStatus) GetData() []PartitionStatusWord {
	count := len(cmd.GetPartitions())
	itemSize := int(cmd.BytesOfStatus)
	return serialization.RemainBytesGetItems(cmd.Statuses, itemSize, count, func(data []byte) PartitionStatusWord {
		bitset := serialization.NewBitMaskFromBytes(data, 0, true).GetBitset()
		return PartitionStatusWord(bitset)
	})
}

func init() {
	registerCommand[PartitionStatus](2066)
}

type PartitionStatusWord []bool

func (word PartitionStatusWord) Armed() bool {
	return word[PartitionStatusArmed]
}

func (word PartitionStatusWord) Stay() bool {
	return word[PartitionStatusStay]
}

func (word PartitionStatusWord) Away() bool {
	return word[PartitionStatusAway]
}

func (word PartitionStatusWord) Night() bool {
	return word[PartitionStatusNight]
}

func (word PartitionStatusWord) NoDelay() bool {
	return word[PartitionStatusNoDelay]
}

func (word PartitionStatusWord) Alarm() bool {
	return word[PartitionStatusAlarm]
}

func (word PartitionStatusWord) Troubles() bool {
	return word[PartitionStatusTroubles]
}

func (word PartitionStatusWord) AlarmInMemory() bool {
	return word[PartitionStatusAlarmInMemory]
}

func (word PartitionStatusWord) Fire() bool {
	return word[PartitionStatusFire]
}

func (word PartitionStatusWord) String() string {
	parts := make([]string, 0)

	if word.Armed() {
		parts = append(parts, "Armed")
	}
	if word.Stay() {
		parts = append(parts, "Stay")
	}
	if word.Away() {
		parts = append(parts, "Away")
	}
	if word.Night() {
		parts = append(parts, "Night")
	}
	if word.NoDelay() {
		parts = append(parts, "NoDelay")
	}
	if word.Alarm() {
		parts = append(parts, "Alarm")
	}
	if word.Troubles() {
		parts = append(parts, "Troubles")
	}
	if word.AlarmInMemory() {
		parts = append(parts, "AlarmInMemory")
	}
	if word.Fire() {
		parts = append(parts, "Fire")
	}

	return "{" + strings.Join(parts, ", ") + "}"
}
