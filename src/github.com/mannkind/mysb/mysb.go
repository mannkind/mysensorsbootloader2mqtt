package main

import (
	"fmt"
	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/mannkind/mysb/controller"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var (
	config  = Config{}
	control = controller.Control{}
	client  *mqtt.Client
)

func initConfig() {
	source, err := ioutil.ReadFile("config.yaml")
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}

	control.SubTopic = config.MQTT.SubTopic
	control.PubTopic = config.MQTT.PubTopic
	control.NextID = config.AutoID.NextID
	control.OTA = config.OTA
}

func initMQTT() {
	log.Println("Connecting to MQTT")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.MQTT.Broker)
	opts.SetClientID(config.MQTT.ClientID)

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Println("Connected to MQTT")

	// Subscribe to topics
	subscriptions := SubscriptionHandler{
		fmt.Sprintf("%s/255/255/3/0/3", config.MQTT.SubTopic): control.IDRequest,
		fmt.Sprintf("%s/+/255/4/0/0", config.MQTT.SubTopic):   control.ConfigurationRequest,
		fmt.Sprintf("%s/+/255/4/0/2", config.MQTT.SubTopic):   control.DataRequest,
		"mysb/+/+": control.BootloaderCommand,
	}

	for topic, handler := range subscriptions {
		if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
}

func main() {
	initConfig()
	initMQTT()

	// Wait forever
	select {}
}
