package main

import (
	"log"
	"reflect"

	"github.com/caarlos0/env"
	mqttExtCfg "github.com/mannkind/paho.mqtt.golang.ext/cfg"
	"gopkg.in/yaml.v2"
)

// Config - Structured configuration for the application.
type Config struct {
	MQTT             *mqttExtCfg.MQTTConfig
	SubTopic         string          `env:"MYSB_SUBTOPIC"         envDefault:"mysensors_rx"`
	PubTopic         string          `env:"MYSB_PUBTOPIC"         envDefault:"mysensors_tx"`
	AutoIDEnabled    bool            `env:"MYSB_AUTOID"           envDefault:"false"`
	NextID           uint            `env:"MYSB_NEXTID"           envDefault:"1"`
	FirmwareBasePath string          `env:"MYSB_FIRMWAREBASEPATH" envDefault:"/config/firmware"`
	Nodes            nodeSettingsMap `env:"MYSB_NODES"`
}

// NewConfig - Returns a new reference to a fully configured object.
func NewConfig(mqttCfg *mqttExtCfg.MQTTConfig) *Config {
	c := Config{}
	c.MQTT = mqttCfg

	if c.MQTT.ClientID == "" {
		c.MQTT.ClientID = "DefaultMysbClientID"
	}

	if err := env.ParseWithFuncs(&c, env.CustomParsers{
		reflect.TypeOf(nodeSettingsMap{}): nodeSettingsParser,
	}); err != nil {
		log.Panicf("Error unmarshaling configuration: %s", err)
	}

	redactedPassword := ""
	if len(c.MQTT.Password) > 0 {
		redactedPassword = "<REDACTED>"
	}

	log.Printf("Environmental Settings:")
	log.Printf("  * ClientID        : %s", c.MQTT.ClientID)
	log.Printf("  * Broker          : %s", c.MQTT.Broker)
	log.Printf("  * SubTopic        : %s", c.SubTopic)
	log.Printf("  * PubTopic        : %s", c.PubTopic)
	log.Printf("  * Username        : %s", c.MQTT.Username)
	log.Printf("  * Password        : %s", redactedPassword)
	log.Printf("  * AutoID          : %t", c.AutoIDEnabled)
	log.Printf("  * NextID          : %d", c.NextID)
	log.Printf("  * FirmwareBasePath: %s", c.FirmwareBasePath)
	log.Printf("  * Nodes           : %+v", c.Nodes)

	return &c
}

func nodeSettingsParser(value string) (interface{}, error) {
	c := make(nodeSettingsMap)
	if err := yaml.Unmarshal([]byte(value), &c); err != nil {
		log.Panicf("Error unmarshaling control configuration: %s", err)
	}

	return c, nil
}
