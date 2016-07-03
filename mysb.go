package main

import (
	"flag"
	"github.com/mannkind/mysb/transport"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

func main() {
	c := flag.String("c", "config.yaml", "/the/path/to/config.yaml")
	flag.Parse()

	filename := *c
	log.Printf("Loading Configuration %s", filename)
	source, rfErr := ioutil.ReadFile(filename)
	if rfErr != nil {
		log.Panicf("Error reading configuration: %s", rfErr)
	}

	mqtt := transport.MQTT{}
	uErr := yaml.Unmarshal(source, &mqtt)
	if uErr != nil {
		log.Panicf("Error unmarshaling configuration: %s", uErr)
	}
	log.Printf("Loaded Configuration %s", filename)

	sErr := mqtt.Start()
	if sErr != nil {
		log.Panicf("Error starting transport.MQTT: %s", sErr)
	}

	select {}
}
