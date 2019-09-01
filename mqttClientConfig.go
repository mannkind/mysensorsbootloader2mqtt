package main

import "github.com/mannkind/twomqtt"

type mqttClientConfig struct {
	globalClientConfig
	MQTTProxyConfig twomqtt.MQTTProxyConfig

	SubTopic           string          `env:"MYSENSORS_SUBTOPIC"         envDefault:"mysensors_rx"`
	PubTopic           string          `env:"MYSENSORS_PUBTOPIC"         envDefault:"mysensors_tx"`
	AutoIDEnabled      bool            `env:"MYSENSORS_AUTOID"           envDefault:"false"`
	NextID             uint            `env:"MYSENSORS_NEXTID"           envDefault:"1"`
	FirmwareBasePath   string          `env:"MYSENSORS_FIRMWAREBASEPATH" envDefault:"/config/firmware"`
	Nodes              nodeSettingsMap `env:"MYSENSORS_NODES"`
	BootloaderCommands bootloaderCmdMap
}
