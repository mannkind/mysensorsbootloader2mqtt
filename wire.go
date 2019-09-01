//+build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/mannkind/twomqtt"
)

func initialize() *mqttClient {
	wire.Build(
		newMQTTClient,
		newConfig,
		wire.FieldsOf(new(config), "MQTTClientConfig"),
		wire.FieldsOf(new(mqttClientConfig), "MQTTProxyConfig"),
		twomqtt.NewMQTTProxy,
	)

	return &mqttClient{}
}
