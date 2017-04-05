package main

import (
	"log"

	"github.com/mannkind/mysb/cmd"
)

// Version - Set during compilation when using included Makefile
var Version = "X.X.X"

func main() {
	log.Printf("Mysb Version: %s", Version)
	cmd.Execute()
}
