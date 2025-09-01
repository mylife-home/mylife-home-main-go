package log

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)

	WithError(err error) Logger
}
