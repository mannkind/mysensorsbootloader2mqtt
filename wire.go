//+build wireinject

package main

import (
	"github.com/google/wire"
	mqttExtCfg "github.com/mannkind/paho.mqtt.golang.ext/cfg"
	mqttExtDI "github.com/mannkind/paho.mqtt.golang.ext/di"
)

// InitializeMysb - Compile-time DI
func InitializeMysb() *Mysb {
	wire.Build(mqttExtCfg.NewMQTTConfig, NewConfig, mqttExtDI.NewMQTTFuncWrapper, NewMysb)

	return &Mysb{}
}
