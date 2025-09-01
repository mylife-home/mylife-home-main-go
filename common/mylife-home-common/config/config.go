package config

import (
	config "github.com/gookit/config/v2"
	yaml "github.com/gookit/config/v2/yaml"

	log "mylife-home-common/log"
)

var logger = log.CreateLogger("mylife:home:config")
var conf *config.Config

func Init(configFile string) {
	conf = config.NewWithOptions("mylife-home-config", config.ParseEnv, config.Readonly)

	// add driver for support yaml content
	conf.AddDriver(yaml.Driver)

	err := conf.LoadFiles(configFile)
	if err != nil {
		panic(err)
	}

	logger.Infof("Config loaded: %+v", conf.Data())
}

func BindStructure(key string, value any) {
	err := conf.Structure(key, value)
	if err != nil {
		panic(err)
	}

	logger.Debugf("Config '%s' fetched: %+v", key, value)
}

func GetString(key string) string {
	value := conf.MustString(key)

	logger.Debugf("Config '%s' fetched: %s", key, value)
	return value
}

func FindString(key string) (string, bool) {
	value, ok := conf.GetValue(key, false)

	if !ok {
		logger.Debugf("Config '%s' not fetched", key)
		return "", false
	}

	str, ok := value.(string)
	if !ok {
		logger.Warnf("Config '%s' wrong value type '%+v'", key, value)
		return "", false
	}

	logger.Debugf("Config '%s' fetched: %s", key, str)
	return str, true
}
