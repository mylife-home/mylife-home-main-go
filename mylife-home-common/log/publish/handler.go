package publish

import (
	"fmt"
	"mylife-home-common/tools"
	"time"

	"github.com/apex/log"
)

var onEntry = tools.NewCallbackManager[*LogEntry]()

type LogLevel string

const (
	Debug LogLevel = "debug"
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
	Error LogLevel = "error"
)

type LogEntry struct {
	timestamp  time.Time
	loggerName string
	level      LogLevel
	message    string
	err        *LogError
}

func (le *LogEntry) Timestamp() time.Time {
	return le.timestamp
}

func (le *LogEntry) LoggerName() string {
	return le.loggerName
}

func (le *LogEntry) Level() LogLevel {
	return le.level
}

func (le *LogEntry) Message() string {
	return le.message
}

func (le *LogEntry) Error() *LogError {
	return le.err
}

type LogError struct {
	message    string
	stacktrace string
}

func (le *LogError) Message() string {
	return le.message
}

func (le *LogError) StackTrace() string {
	return le.stacktrace
}

func OnEntry() tools.CallbackRegistration[*LogEntry] {
	return onEntry
}

// Handler implementation.
type Handler struct {
}

// New handler.
func New() *Handler {
	return &Handler{}
}

func (h *Handler) HandleLog(e *log.Entry) error {
	entry := &LogEntry{
		timestamp:  e.Timestamp,
		loggerName: e.Fields.Get("logger-name").(string),
		level:      convertLevel(e.Level),
		message:    e.Message,
		err:        convertError(e.Fields.Get("error")),
	}

	onEntry.Execute(entry)

	return nil
}

func convertLevel(level log.Level) LogLevel {
	switch level {
	case log.DebugLevel:
		return Debug
	case log.InfoLevel:
		return Info
	case log.WarnLevel:
		return Warn
	case log.ErrorLevel, log.FatalLevel:
		return Error
	}

	panic(fmt.Errorf("unknown level %d", level))
}

func convertError(err interface{}) *LogError {
	if err == nil {
		return nil
	}

	terr := err.(error)

	return &LogError{
		message:    terr.Error(),
		stacktrace: tools.GetStackTraceStr(terr),
	}
}
