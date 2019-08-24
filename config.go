package main

import (
	"reflect"

	"github.com/caarlos0/env/v6"
	mqttExtCfg "github.com/mannkind/paho.mqtt.golang.ext/cfg"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config - Structured configuration for the application.
type Config struct {
	MQTT             *mqttExtCfg.MQTTConfig
	SubTopic         string          `env:"MYSENSORS_SUBTOPIC"         envDefault:"mysensors_rx"`
	PubTopic         string          `env:"MYSENSORS_PUBTOPIC"         envDefault:"mysensors_tx"`
	AutoIDEnabled    bool            `env:"MYSENSORS_AUTOID"           envDefault:"false"`
	NextID           uint            `env:"MYSENSORS_NEXTID"           envDefault:"1"`
	FirmwareBasePath string          `env:"MYSENSORS_FIRMWAREBASEPATH" envDefault:"/config/firmware"`
	Nodes            nodeSettingsMap `env:"MYSENSORS_NODES"`
	DebugLogLevel    bool            `env:"MYSENSORS_DEBUG" envDefault:"false"`
}

func newConfig(mqttCfg *mqttExtCfg.MQTTConfig) *Config {
	c := Config{}
	c.MQTT = mqttCfg

	if err := env.ParseWithFuncs(&c, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(nodeSettingsMap{}): nodeSettingsParser,
	}); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Unable to unmarshal configuration")
	}

	c.MQTT.Defaults("DefaultMySensorsBootloaderClientID", "", "")

	log.WithFields(log.Fields{
		"MySensors.AutoIDEnabled":    c.AutoIDEnabled,
		"MySensors.SubTopic":         c.SubTopic,
		"MySensors.PubTopic":         c.PubTopic,
		"MySensors.NextID":           c.NextID,
		"MySensors.Nodes":            c.Nodes,
		"MySensors.FirmwareBasePath": c.FirmwareBasePath,
		"MySensors.DebugLogLevel":    c.DebugLogLevel,
	}).Info("Service Environmental Settings")

	if c.DebugLogLevel {
		log.SetLevel(log.DebugLevel)
		log.Debug("Enabling the debug log level")
	}

	return &c
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
