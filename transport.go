package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/eclipse/paho.mqtt.golang"
)

// Go-ified MySensors Constants
const (
	idRequestTopic                 = "%s/255/255/3/0/3"
	idResponseTopic                = "%s/255/255/3/0/4"
	firmwareConfigRequestTopic     = "%s/+/255/4/0/0"
	firmwareConfigResponseTopic    = "%s/%s/255/4/0/1"
	firmwareRequestTopic           = "%s/+/255/4/0/2"
	firmwareResponseTopic          = "%s/%s/255/4/0/3"
	firmwareBootloaderCommandTopic = "mysensors/bootloader/+/+"
)

// MysbMQTT - MQTT all the things!
type MysbMQTT struct {
	Client   mqtt.Client
	Settings struct {
		ClientID string
		Broker   string
		SubTopic string
		PubTopic string
		Username string
		Password string
	}
	Control       Control
	LastPublished string
}

// Start - Connect and Subscribe
func (t *MysbMQTT) Start() error {
	log.Println("Connecting to MQTT: ", t.Settings.Broker)
	opts := mqtt.NewClientOptions().
		AddBroker(t.Settings.Broker).
		SetClientID(t.Settings.ClientID).
		SetOnConnectHandler(t.onConnect).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			log.Printf("Disconnected from MQTT: %s.", err)
		}).
		SetUsername(t.Settings.Username).
		SetPassword(t.Settings.Password)

	t.Client = mqtt.NewClient(opts)
	if token := t.Client.Connect(); !token.Wait() || token.Error() != nil {
		return token.Error()
	}

	return nil
}

// Stop - Disconnect
func (t *MysbMQTT) Stop() {
	if t.Client != nil && t.Client.IsConnected() {
		t.Client.Disconnect(0)
	}
}

func (t *MysbMQTT) onConnect(client mqtt.Client) {
	log.Println("Connected to MQTT")

	// Subscribe to topics
	subscriptions := map[string]mqtt.MessageHandler{
		fmt.Sprintf(idRequestTopic, t.Settings.SubTopic):             t.idRequest,
		fmt.Sprintf(firmwareConfigRequestTopic, t.Settings.SubTopic): t.configurationRequest,
		fmt.Sprintf(firmwareRequestTopic, t.Settings.SubTopic):       t.dataRequest,
		firmwareBootloaderCommandTopic:                               t.bootloaderCommand,
	}

	//
	if !client.IsConnected() {
		log.Print("Subscribe Error: Not Connected (Reloading Config?)")
		return
	}

	for topic, handler := range subscriptions {
		log.Printf("Subscribing: %s", topic)
		if token := client.Subscribe(topic, 0, handler); !token.Wait() || token.Error() != nil {
			log.Printf("Subscribe Error: %s", token.Error())
		}
	}
}

func (t *MysbMQTT) idRequest(client mqtt.Client, msg mqtt.Message) {
	if newID, incremented := t.Control.IDRequest(); incremented {
		t.publish(client, fmt.Sprintf(idResponseTopic, t.Settings.PubTopic), newID)
	}
}

func (t *MysbMQTT) configurationRequest(client mqtt.Client, msg mqtt.Message) {
	_, payload, to := t.msgParts(msg)

	// Attempt to run any bootloader commands
	if t.runBootloaderCommand(client, to) {
		return
	}

	t.publish(
		client,
		fmt.Sprintf(firmwareConfigResponseTopic, t.Settings.PubTopic, to),
		t.Control.ConfigurationRequest(to, payload),
	)
}

func (t *MysbMQTT) dataRequest(client mqtt.Client, msg mqtt.Message) {
	_, payload, to := t.msgParts(msg)

	t.publish(
		client,
		fmt.Sprintf(firmwareResponseTopic, t.Settings.PubTopic, to),
		t.Control.DataRequest(to, payload),
	)
}

func (t *MysbMQTT) bootloaderCommand(client mqtt.Client, msg mqtt.Message) {
	topic, payload, _ := t.msgParts(msg)

	parts := strings.Split(topic, "/")
	to := parts[2]
	cmd := parts[3]

	t.Control.BootloaderCommand(to, cmd, payload)
}

func (t *MysbMQTT) runBootloaderCommand(client mqtt.Client, to string) bool {
	if blcmd, ok := t.Control.BootloaderCommands[to]; ok {
		outTopic := fmt.Sprintf(firmwareConfigResponseTopic, t.Settings.PubTopic, to)
		outPayload := blcmd.String()
		t.publish(client, outTopic, outPayload)

		delete(t.Control.BootloaderCommands, to)
		return true
	}

	return false
}

func (t *MysbMQTT) msgParts(msg mqtt.Message) (string, string, string) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	return topic, payload, to
}

func (t *MysbMQTT) publish(client mqtt.Client, topic string, payload string) {
	if token := client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
		log.Printf("Publish Error: %s", token.Error())
	}
	t.LastPublished = fmt.Sprintf("%s %s", topic, payload)
}
