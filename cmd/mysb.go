package cmd

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/mannkind/mysb/controller"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			controller := controller.MysbMQTT{}
			if err := viper.Unmarshal(&controller); err != nil {
				log.Panicf("Error unmarshaling configuration: %s", err)
			}

			if err := controller.Start(); err != nil {
				log.Panicf("Error starting controller: %s", err)
			}

			<-reload
			controller.Stop()
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
