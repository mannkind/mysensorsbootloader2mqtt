package main

import (
	"log"
	"reflect"

	"github.com/caarlos0/env"
	"gopkg.in/yaml.v2"
)

// Version - Set during compilation when using included Makefile
var Version = "X.X.X"

func main() {
	log.Printf("Mysb Version: %s", Version)

	log.Print("Stating Process")
	controller := mysb{}
	if err := env.ParseWithFuncs(&controller, env.CustomParsers{
		reflect.TypeOf(nodeSettingsMap{}): nodeSettingsParser,
	}); err != nil {
		log.Panicf("Error unmarshaling configuration: %s", err)
	}

	redactedPassword := ""
	if len(controller.Password) > 0 {
		redactedPassword = "<REDACTED>"
	}

	log.Printf("Environmental Settings:")
	log.Printf("  * ClientID      : %s", controller.ClientID)
	log.Printf("  * Broker        : %s", controller.Broker)
	log.Printf("  * SubTopic      : %s", controller.SubTopic)
	log.Printf("  * PubTopic      : %s", controller.PubTopic)
	log.Printf("  * Username      : %s", controller.Username)
	log.Printf("  * Password      : %s", redactedPassword)
	log.Printf("  * AutoID          : %t", controller.AutoIDEnabled)
	log.Printf("  * NextID          : %d", controller.NextID)
	log.Printf("  * FirmwareBasePath: %s", controller.FirmwareBasePath)
	log.Printf("  * Nodes           : %+v", controller.Nodes)

	if err := controller.start(); err != nil {
		log.Panicf("Error starting mysb: %s", err)
	}

	// log.Print("Ending Process")
	select {}
}

func nodeSettingsParser(value string) (interface{}, error) {
	control := make(nodeSettingsMap)
	if err := yaml.Unmarshal([]byte(value), &control); err != nil {
		log.Panicf("Error unmarshaling control configuration: %s", err)
	}

	return control, nil
}
