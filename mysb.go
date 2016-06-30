package main

import (
	"flag"
	"github.com/fsnotify/fsnotify"
	"github.com/mannkind/mysb/transport"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"strings"
)

var (
	mqtt transport.MQTT
)

func main() {
	c := flag.String("c", "config.yaml", "/the/path/to/config.yaml")
	flag.Parse()

	filename := *c
	dir, file := filepath.Split(filename)
	file = strings.Replace(file, filepath.Ext(file), "", -1)
	if dir == "" {
		dir = "."
	}

	viper.SetConfigName(file)
	viper.AddConfigPath(dir)
	viper.WatchConfig()

	mqtt = transport.MQTT{}
	log.Printf("Loading Configuration %s", filename)
	if err := viper.ReadInConfig(); err != nil {
		log.Print("Unable to read configuration file")
		panic(err)
	}

	if err := viper.Unmarshal(&mqtt); err != nil {
		log.Printf("Unable to load configuration; %s", err)
		panic(err)
	}
	log.Printf("Loaded Configuration %s", filename)

	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Reloading Configuration")
		if err := viper.Unmarshal(&mqtt); err != nil {
			log.Printf("Unable to load configuration; %s", err)
			panic(err)
		}
		log.Println("Reloaded Configuration")
		mqtt.Restart()
	})

	mqtt.Start()
	select {}
}
