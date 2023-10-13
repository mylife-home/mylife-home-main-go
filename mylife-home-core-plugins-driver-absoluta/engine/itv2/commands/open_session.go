package commands

var _ Command = (*OpenSession)(nil)
var _ CommandWithAppSeq = (*OpenSession)(nil)

type OpenSession struct {
	AppSeq uint8

	DeviceTypeOrVendorID uint8
	DeviceId             uint16
	SoftwareVersion      [2]byte // BCD string
	ProtocolVersion      [2]byte // BCD string
	TxSize               uint16
	RxSize               uint16
	Unused               uint16 // set to 1
	EncryptionType       uint8  // 0 or 1
}

func init() {
	registerCommand[OpenSession](1546)
}

func (cmd *OpenSession) GetAppSeq() uint8 {
	return cmd.AppSeq
}

func (cmd *OpenSession) SetAppSeq(value uint8) {
	cmd.AppSeq = value
}
