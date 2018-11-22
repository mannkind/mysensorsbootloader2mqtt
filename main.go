package main

import (
	"log"
)

// Version - Set during compilation when using included Makefile
var Version = "X.X.X"

func main() {
	log.Printf("Mysb Version: %s", Version)

	m := InitializeMysb()
	if err := m.Run(); err != nil {
		log.Panicf("Error starting mysb process: %s", err)
	}

	select {}
}
