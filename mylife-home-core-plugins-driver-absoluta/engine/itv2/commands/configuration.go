package commands

import (
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"

	"golang.org/x/text/encoding/charmap"
)

const (
	ConfigurationOptionZoneLabel                        uint8 = 209
	ConfigurationOptionPartitionLabel                   uint8 = 211
	ConfigurationOptionUserLabel                        uint8 = 217
	ConfigurationOptionKeypadLabel                      uint8 = 226
	ConfigurationOptionZoneExpanderModuleLabel          uint8 = 227
	ConfigurationOptionPowerSupplyModuleLabel           uint8 = 228
	ConfigurationOptionHighCurrentOutputModuleLabel     uint8 = 229
	ConfigurationOptionOutputExpanderModuleLabel        uint8 = 230
	ConfigurationOptionWirelessTransceiverLabel         uint8 = 231
	ConfigurationOptionAlternateCommunicatorModuleLabel uint8 = 232
	ConfigurationOptionWirelessSirenLabel               uint8 = 233
	ConfigurationOptionWirelessRepeaterLabel            uint8 = 234
	ConfigurationOptionAudioVerificationModuleLabel     uint8 = 235
	ConfigurationOptionAbsolutaZoneLabel                uint8 = 1
	ConfigurationOptionAbsolutaPartitionLabel           uint8 = 3
	ConfigurationOptionAbsolutaCommandLabels            uint8 = 4
	ConfigurationOptionAbsolutaArmingModeLabel          uint8 = 13
)

var _ RequestData = (*ConfigurationRequest)(nil)
var _ ResponseData = (*Configuration)(nil)

type ConfigurationRequest struct {
	OptionId           *serialization.VarBytes
	OptionIdOffsetFrom *serialization.VarBytes
	OptionIdOffsetTo   *serialization.VarBytes // TODO: does not exist if OptionIdOffsetFrom is null
}

func (req *ConfigurationRequest) RequestCode() uint16 {
	return 1905
}

type Configuration struct {
	Req        ConfigurationRequest
	DataLength *serialization.VarBytes
	Data       *serialization.RemainBytes // TODO: should be an array. each elem size = DataLength,
}

func (cmd *Configuration) GetRequest() RequestData {
	return &cmd.Req
}
func (cmd *Configuration) Count() int {
	if cmd.Req.OptionIdOffsetFrom.IsNull() || cmd.Req.OptionIdOffsetTo.IsNull() {
		return 1
	} else {
		return int(cmd.Req.OptionIdOffsetTo.GetUint()) - int(cmd.Req.OptionIdOffsetFrom.GetUint()) + 1
	}
}

func (cmd *Configuration) GetItemSize() int {
	return int(cmd.DataLength.GetUint())
}

func (cmd *Configuration) GetItem(index int) []byte {
	itemSize := cmd.GetItemSize()
	begin := index * itemSize
	end := begin + itemSize
	return cmd.Data.Get()[begin:end]
}

func (cmd *Configuration) GetStrings(encoding *charmap.Charmap) []string {
	array := make([]string, 0, cmd.Count())

	for index := 0; index < cmd.Count(); index += 1 {
		item := cmd.GetItem(index)

		converted, err := encoding.NewDecoder().Bytes(item)
		if err != nil {
			logger.Warnf("Could not convert string from 'Windows1251'")
			converted = item
		}

		array = append(array, string(converted))
	}

	return array
}

func init() {
	registerCommand[Configuration](1905)
}
