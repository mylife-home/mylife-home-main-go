package log

import "github.com/apex/log"

var _ Logger = (*loggerImpl)(nil)

type loggerImpl struct {
	impl log.Interface
}

func newLoggerImpl(impl log.Interface) *loggerImpl {
	return &loggerImpl{impl}
}

func (logger *loggerImpl) Debugf(format string, args ...interface{}) {
	logger.impl.Debugf(format, args...)
}

func (logger *loggerImpl) Infof(format string, args ...interface{}) {
	logger.impl.Infof(format, args...)
}

func (logger *loggerImpl) Warnf(format string, args ...interface{}) {
	logger.impl.Warnf(format, args...)
}

func (logger *loggerImpl) Errorf(format string, args ...interface{}) {
	logger.impl.Errorf(format, args...)
}

func (logger *loggerImpl) Debug(msg string) {
	logger.impl.Debug(msg)
}

func (logger *loggerImpl) Info(msg string) {
	logger.impl.Info(msg)
}

func (logger *loggerImpl) Warn(msg string) {
	logger.impl.Warn(msg)
}

func (logger *loggerImpl) Error(msg string) {
	logger.impl.Error(msg)
}

func (logger *loggerImpl) WithError(err error) Logger {
	return newLoggerImpl(logger.impl.WithField("error", err))
}
