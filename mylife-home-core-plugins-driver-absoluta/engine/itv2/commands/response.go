package commands

import "fmt"

type ResponseCode = byte

const (
	ResponseCodeSuccess                           ResponseCode = 0
	ResponseCodeInvalidIdentifier                 ResponseCode = 1
	ResponseCodeUnsupportedEncryptionType         ResponseCode = 3
	ResponseCodeInvalidAccessCode                 ResponseCode = 17
	ResponseCodeInvalidPartition                  ResponseCode = 20
	ResponseCodeNoTroublesPresentForRequestedType ResponseCode = 26
	ResponseCodeNoRequestedAlarmsFound            ResponseCode = 27
)

type Response struct {
	CommandSeq byte
	Code       ResponseCode
}

func (cmd *Response) CodeString() string {
	switch cmd.Code {
	case ResponseCodeSuccess:
		return "Success"
	case ResponseCodeInvalidIdentifier:
		return "InvalidIdentifier"
	case ResponseCodeUnsupportedEncryptionType:
		return "UnsupportedEncryptionType"
	case ResponseCodeInvalidAccessCode:
		return "InvalidAccessCode"
	case ResponseCodeInvalidPartition:
		return "InvalidPartition"
	case ResponseCodeNoTroublesPresentForRequestedType:
		return "NoTroublesPresentForRequestedType"
	case ResponseCodeNoRequestedAlarmsFound:
		return "NoRequestedAlarmsFound"
	default:
		return fmt.Sprintf("<%d>", cmd.Code)
	}
}

func init() {
	registerCommand[Response](1282)
}
