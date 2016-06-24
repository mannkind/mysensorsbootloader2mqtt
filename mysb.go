package main

import "github.com/mannkind/mysb/transport"

func main() {
	mqtt := *transport.NewMQTT("config.yaml")
	mqtt.Loop()
}
