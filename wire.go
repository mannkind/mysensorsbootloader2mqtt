//+build wireinject

package main

import (
	"github.com/google/wire"
	mqttExtCfg "github.com/mannkind/paho.mqtt.golang.ext/cfg"
	mqttExtDI "github.com/mannkind/paho.mqtt.golang.ext/di"
)

func initialize() *mqttClient {
	wire.Build(mqttExtCfg.NewMQTTConfig, NewConfig, mqttExtDI.NewMQTTFuncWrapper, newMqttClient)

	return &mqttClient{}
}
