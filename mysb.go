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

	mqtt := transport.MQTT{}

	filename := *c
	log.Printf("Loading Configuration %s", filename)
	source, err := ioutil.ReadFile(filename)
	err = yaml.Unmarshal(source, &mqtt)
	if err != nil {
		panic(err)
	}
	log.Printf("Loaded Configuration %s", filename)

	mqtt.Start()
	select {}
}
