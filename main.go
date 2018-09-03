package main

import (
	"log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var reload = make(chan bool)

// Version - Set during compilation when using included Makefile
var Version = "X.X.X"

// MysbCmd - The root Mysb commands
var MysbCmd = &cobra.Command{
	Use:   "mysb",
	Short: "A Firmware Uploading Tool for the MYSBootloader via MQTT",
	Long:  "A Firmware Uploading Tool for the MYSBootloader via MQTT",
	Run: func(cmd *cobra.Command, args []string) {
		for {
			log.Printf("Creating the MQTT transport handler")
			controller := MysbMQTT{}
			if err := viper.Unmarshal(&controller); err != nil {
				log.Panicf("Error unmarshaling configuration: %s", err)
			}

			if err := controller.Start(); err != nil {
				log.Panicf("Error starting MQTT transport handler: %s", err)
			}

			<-reload
			log.Printf("Received Reload Signal")
			controller.Stop()
		}
	},
}

func init() {
	cobra.OnInitialize(func() {
		viper.SetConfigFile(cfgFile)
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("Configuration Changed: %s", e.Name)
			reload <- true
		})

		log.Printf("Loading Configuration %s", cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error Loading Configuration: %s ", err)
		}
		log.Printf("Loaded Configuration %s", cfgFile)
	})

	MysbCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", ".mysb.yaml", "The path to the configuration file")
}

func main() {
	log.Printf("Mysb Version: %s", Version)
	Execute()
}

// Execute - Adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := MysbCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}