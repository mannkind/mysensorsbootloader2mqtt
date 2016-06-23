package main

import (
	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

// SubscriptionHandler - Mapping topic to handler
type SubscriptionHandler map[string]mqtt.MessageHandler

// Config - Config all the things!
type Config struct {
	MQTT struct {
		ClientID string
		Broker   string
		SubTopic string
		PubTopic string
	}
	AutoID struct {
		NextID uint8
	}
	OTA struct {
		Types    map[string]string
		Versions map[string]string
		Firmware map[string]map[string]string
		Nodes    map[string]map[string]string
	}
}
