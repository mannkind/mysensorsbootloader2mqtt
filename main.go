package main

import (
	"log"

	"go.uber.org/dig"
)

// Version - Set during compilation when using included Makefile
var Version = "X.X.X"

func main() {
	log.Printf("Mysb Version: %s", Version)

	c := buildContainer()
	err := c.Invoke(func(m *Mysb) error {
		return m.Run()
	})

	if err != nil {
		log.Panicf("Error starting mysb process: %s", err)
	}

	select {}
}

func buildContainer() *dig.Container {
	c := dig.New()
	c.Provide(NewConfig)
	c.Provide(NewMQTTFuncWrapper)
	c.Provide(NewMysb)

	return c
}
