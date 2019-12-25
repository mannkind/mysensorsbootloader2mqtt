//+build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/mannkind/twomqtt"
)

func initialize() *sink {
	wire.Build(
		newOpts,
		newSink,
		wire.FieldsOf(new(sinkOpts), "MQTTOpts"),
		wire.FieldsOf(new(opts), "Sink"),
		twomqtt.NewMQTT,
	)

	return &sink{}
}
