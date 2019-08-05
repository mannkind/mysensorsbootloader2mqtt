package main

import (
	"reflect"

	"github.com/caarlos0/env"
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

// NewConfig - Returns a new reference to a fully configured object.
func NewConfig(mqttCfg *mqttExtCfg.MQTTConfig) *Config {
	c := Config{}
	c.MQTT = mqttCfg

	if c.MQTT.ClientID == "" {
		c.MQTT.ClientID = "DefaultMySensorsBootloaderClientID"
	}

	if err := env.ParseWithFuncs(&c, env.CustomParsers{
		reflect.TypeOf(nodeSettingsMap{}): nodeSettingsParser,
	}); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Unable to unmarshal configuration")
	}

	redactedPassword := ""
	if len(c.MQTT.Password) > 0 {
		redactedPassword = "<REDACTED>"
	}

	log.WithFields(log.Fields{
		"MQTT.ClientID":              c.MQTT.ClientID,
		"MQTT.Broker":                c.MQTT.Broker,
		"MQTT.Username":              c.MQTT.Username,
		"MQTT.Password":              redactedPassword,
		"MQTT.Discovery":             c.MQTT.Discovery,
		"MQTT.DiscoveryPrefix":       c.MQTT.DiscoveryPrefix,
		"MQTT.DiscoveryName":         c.MQTT.DiscoveryName,
		"MQTT.TopicPrefix":           c.MQTT.TopicPrefix,
		"MySensors.AutoIDEnabled":    c.AutoIDEnabled,
		"MySensors.SubTopic":         c.SubTopic,
		"MySensors.PubTopic":         c.PubTopic,
		"MySensors.NextID":           c.NextID,
		"MySensors.Nodes":            c.Nodes,
		"MySensors.FirmwareBasePath": c.FirmwareBasePath,
		"MySensors.DebugLogLevel":    c.DebugLogLevel,
	}).Info("Environmental Settings")

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
