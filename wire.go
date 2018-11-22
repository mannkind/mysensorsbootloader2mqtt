//+build wireinject

package main

import (
	"github.com/google/go-cloud/wire"
)

// InitializeMysb - Compile-time DI
func InitializeMysb() *Mysb {
	wire.Build(NewConfig, NewMQTTFuncWrapper, NewMysb)

	return &Mysb{}
}
