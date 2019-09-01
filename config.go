package main

import (
	"reflect"

	"github.com/caarlos0/env/v6"
	"github.com/mannkind/twomqtt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type config struct {
	GeneralConfig      twomqtt.GeneralConfig
	GlobalClientConfig globalClientConfig
	MQTTClientConfig   mqttClientConfig
}

func newConfig() config {
	c := config{
		GeneralConfig:      twomqtt.GeneralConfig{},
		GlobalClientConfig: globalClientConfig{},
		MQTTClientConfig:   mqttClientConfig{},
	}

	// Manually parse the address:name mapping
	if err := env.ParseWithFuncs(&c, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(nodeSettingsMap{}): nodeSettingsParser,
	}); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Unable to unmarshal configuration")
	}

	if c.GeneralConfig.DebugLogLevel {
		log.SetLevel(log.DebugLevel)
	}

	return c
}

func nodeSettingsParser(value string) (interface{}, error) {
	c := make(nodeSettingsMap)
	if err := yaml.Unmarshal([]byte(value), &c); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Panic("Unable to unmarshal configuration")
	}

	return c, nil
}
