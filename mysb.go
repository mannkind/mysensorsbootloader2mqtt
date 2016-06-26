package main

import (
	"flag"
	"github.com/mannkind/mysb/transport"
)

func main() {
	c := flag.String("c", "config.yaml", "/the/path/to/config.yaml")
	flag.Parse()

	mqtt := *transport.NewMQTT(*c)
	mqtt.ConSub()

	select {}
}
