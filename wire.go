//+build wireinject

package main

import (
	"github.com/google/wire"
)

// InitializeMysb - Compile-time DI
func InitializeMysb() *Mysb {
	wire.Build(NewConfig, NewMQTTFuncWrapper, NewMysb)

	return &Mysb{}
}
