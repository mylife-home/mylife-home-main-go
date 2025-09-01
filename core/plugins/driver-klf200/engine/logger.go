package engine

import (
	"mylife-home-common/log"

	"github.com/mylife-home/klf200-go"
)

var logger = log.CreateLogger("mylife:home:core:plugins:driver-klf200:engine")

var klf200logger = &logWrapper{logger}

type logWrapper struct {
	logger log.Logger
}

func (l *logWrapper) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *logWrapper) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *logWrapper) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *logWrapper) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *logWrapper) Debug(msg string) {
	l.logger.Debug(msg)
}

func (l *logWrapper) Info(msg string) {
	l.logger.Info(msg)
}

func (l *logWrapper) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *logWrapper) Error(msg string) {
	l.logger.Error(msg)
}

func (l *logWrapper) WithError(err error) klf200.Logger {
	return &logWrapper{l.logger.WithError(err)}
}
