package bus

import (
	"fmt"
	"mylife-home-common/log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func init() {
	logger := log.CreateLogger("paho:mqtt")

	mqtt.ERROR = &mqttErrorLogger{logger}
	mqtt.CRITICAL = &mqttErrorLogger{logger}
	if false {
		// Too verbose
		mqtt.WARN = &mqttWarnLogger{logger}
		mqtt.DEBUG = &mqttDebugLogger{logger}
	}
}

type mqttDebugLogger struct {
	impl log.Logger
}

func (ml *mqttDebugLogger) Println(v ...interface{}) {
	ml.impl.Debug(fmt.Sprint(v...))
}

func (ml *mqttDebugLogger) Printf(format string, v ...interface{}) {
	ml.impl.Debugf(format, v...)
}

type mqttWarnLogger struct {
	impl log.Logger
}

func (ml *mqttWarnLogger) Println(v ...interface{}) {
	ml.impl.Warn(fmt.Sprint(v...))
}

func (ml *mqttWarnLogger) Printf(format string, v ...interface{}) {
	ml.impl.Warnf(format, v...)
}

type mqttErrorLogger struct {
	impl log.Logger
}

func (ml *mqttErrorLogger) Println(v ...interface{}) {
	ml.impl.Error(fmt.Sprint(v...))
}

func (ml *mqttErrorLogger) Printf(format string, v ...interface{}) {
	ml.impl.Errorf(format, v...)
}
