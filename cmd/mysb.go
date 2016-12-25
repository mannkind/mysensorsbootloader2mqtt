package cmd

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mannkind/mysb/transport"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const version string = "0.4.1"

var cfgFile string
var reload = make(chan bool)

// MysbCmd - The root Mysb commands
var MysbCmd = &cobra.Command{
	Use:   "mysb",
	Short: "A Firmware Uploading Tool for the MYSBootloader via MQTT",
	Long:  "A Firmware Uploading Tool for the MYSBootloader via MQTT",
	Run: func(cmd *cobra.Command, args []string) {
		for {
			mqtt := transport.MQTT{}
			if err := viper.Unmarshal(&mqtt); err != nil {
				log.Panicf("Error unmarshaling configuration: %s", err)
			}
			mqtt.Control.AutoIDEnabled = len(viper.GetString("control.nextid")) != 0

			if err := mqtt.Start(); err != nil {
				log.Panicf("Error starting transport.MQTT: %s", err)
			}

			<-reload
			if mqtt.Client != nil && mqtt.Client.IsConnected() {
				mqtt.Client.Disconnect(0)
				time.Sleep(500 * time.Millisecond)
			}
		}
	},
}

// Execute - Adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := MysbCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	log.Printf("Mysb Version: %s", version)

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
