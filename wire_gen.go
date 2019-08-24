// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/mannkind/paho.mqtt.golang.ext/cfg"
	"github.com/mannkind/paho.mqtt.golang.ext/client"
)

// Injectors from wire.go:

func initialize() *mqttClient {
	mqttConfig := cfg.NewMQTTConfig()
	config := newConfig(mqttConfig)
	mqttClientWrapper := di.NewMQTTClientWrapper(mqttConfig)
	mainMqttClient := newMQTTClient(config, mqttClientWrapper)
	return mainMqttClient
}
