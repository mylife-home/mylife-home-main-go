package commands

import "fmt"

const (
	ErrorCodeInvalidArgument byte = 1
	ErrorCodeUnknownCommand  byte = 2
	ErrorCodeWebAppConnected byte = 9
)

type Error struct {
	ReceivedCommand uint16
	ErrorCode       byte
}

func (cmd *Error) ErrorCodeString() string {
	switch cmd.ErrorCode {
	case ErrorCodeInvalidArgument:
		return "InvalidArgument"
	case ErrorCodeUnknownCommand:
		return "UnknownCommand"
	case ErrorCodeWebAppConnected:
		return "WebAppConnected"
	default:
		return fmt.Sprintf("<%d>", int(cmd.ErrorCode))
	}
}

func init() {
	registerCommand[Error](1281)
}
