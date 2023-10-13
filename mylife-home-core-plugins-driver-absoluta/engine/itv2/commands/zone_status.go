package commands

import (
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"
	"strings"
)

var _ RequestData = (*ZoneStatusRequest)(nil)
var _ ResponseData = (*ZoneStatus)(nil)

const (
	// Note: this are indexes on status words
	ZoneStatusOpen          int = 0 // = active
	ZoneStatusTamper        int = 1
	ZoneStatusFault         int = 2
	ZoneStatusLowBattery    int = 3
	ZoneStatusDelinquency   int = 4
	ZoneStatusAlarm         int = 5
	ZoneStatusAlarmInMemory int = 6
	ZoneStatusByPassed      int = 7
)

type ZoneStatusRequest struct {
	ZoneNumber    *serialization.VarBytes
	NumberOfZones *serialization.VarBytes
}

func (req *ZoneStatusRequest) RequestCode() uint16 {
	return 2065
}

type ZoneStatus struct {
	Req                 ZoneStatusRequest
	LengthOfStatusBytes uint8
	ZoneStatuses        *serialization.RemainBytes // TODO: DscArray<DscBitMask>, array size = number of zone, bitmask size = lengthOfStatusBytes
}

func (cmd *ZoneStatus) GetRequest() RequestData {
	return &cmd.Req
}

func (cmd *ZoneStatus) GetData() []ZoneStatusWord {
	count := int(cmd.Req.NumberOfZones.GetUint())
	itemSize := int(cmd.LengthOfStatusBytes)
	return serialization.RemainBytesGetItems(cmd.ZoneStatuses, itemSize, count, func(data []byte) ZoneStatusWord {
		bitset := serialization.NewBitMaskFromBytes(data, 0, true).GetBitset()
		return ZoneStatusWord(bitset)
	})
}

func init() {
	registerCommand[ZoneStatus](2065)
}

type ZoneStatusWord []bool

// = active
func (word ZoneStatusWord) Open() bool {
	return word[ZoneStatusOpen]
}

func (word ZoneStatusWord) Tamper() bool {
	return word[ZoneStatusTamper]
}

func (word ZoneStatusWord) Fault() bool {
	return word[ZoneStatusFault]
}

func (word ZoneStatusWord) LowBattery() bool {
	return word[ZoneStatusLowBattery]
}

func (word ZoneStatusWord) Delinquency() bool {
	return word[ZoneStatusDelinquency]
}

func (word ZoneStatusWord) Alarm() bool {
	return word[ZoneStatusAlarm]
}

func (word ZoneStatusWord) AlarmInMemory() bool {
	return word[ZoneStatusAlarmInMemory]
}

func (word ZoneStatusWord) ByPassed() bool {
	return word[ZoneStatusByPassed]
}

func (word ZoneStatusWord) String() string {
	parts := make([]string, 0)

	if word.Open() {
		parts = append(parts, "Open")
	}
	if word.Tamper() {
		parts = append(parts, "Tamper")
	}
	if word.Fault() {
		parts = append(parts, "Fault")
	}
	if word.LowBattery() {
		parts = append(parts, "LowBattery")
	}
	if word.Delinquency() {
		parts = append(parts, "Delinquency")
	}
	if word.Alarm() {
		parts = append(parts, "Alarm")
	}
	if word.AlarmInMemory() {
		parts = append(parts, "AlarmInMemory")
	}
	if word.ByPassed() {
		parts = append(parts, "ByPassed")
	}

	return "{" + strings.Join(parts, ", ") + "}"
}
