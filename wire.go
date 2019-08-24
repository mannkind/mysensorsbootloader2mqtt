//+build wireinject

package main

import (
	"github.com/google/wire"
	mqttExtCfg "github.com/mannkind/paho.mqtt.golang.ext/cfg"
	mqttExtClient "github.com/mannkind/paho.mqtt.golang.ext/client"
)

func initialize() *mqttClient {
	wire.Build(mqttExtCfg.NewMQTTConfig, mqttExtClient.NewMQTTClientWrapper, newConfig, newMQTTClient)

	return &mqttClient{}
}
