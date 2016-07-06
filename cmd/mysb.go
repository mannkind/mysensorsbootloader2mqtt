package cmd

import (
	"github.com/fsnotify/fsnotify"
	"github.com/mannkind/mysb/transport"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"time"
)

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

			if err := mqtt.Start(); err != nil {
				log.Panicf("Error starting transport.MQTT: %s", err)
			}

			select {
			case <-reload:
				if mqtt.Client != nil && mqtt.Client.IsConnected() {
					mqtt.Client.Disconnect(0)
					time.Sleep(500 * time.Millisecond)
				}
				continue
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