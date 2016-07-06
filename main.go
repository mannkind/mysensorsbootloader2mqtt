package main

import (
	"github.com/mannkind/mysb/cmd"
	"log"
)

func main() {
	if err := cmd.MysbCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
