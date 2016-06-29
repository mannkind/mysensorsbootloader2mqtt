package transport

import (
	"fmt"
	"github.com/mannkind/mysb/ota"
	"github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

// MQTT - MQTT all the things!
type MQTT struct {
	Settings struct {
		ClientID string
		Broker   string
		SubTopic string
		PubTopic string
	}
	Control       ota.Control
	LastPublished string
}

// NewMQTT - Create a new MQTT
func NewMQTT(filename string) *MQTT {
	log.Printf("Loading configuration file: %s", filename)

	config := MQTT{}
	source, err := ioutil.ReadFile(filename)
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}
	log.Println("Configuration file loaded")

	return &config
}

// Start - Connect and Subscribe
func (t *MQTT) Start() error {
	log.Print("Connecting to MQTT... ")
	opts := mqtt.NewClientOptions().
		AddBroker(t.Settings.Broker).
		SetClientID(t.Settings.ClientID).
        SetOnConnectHandler(t.onConnect).
        SetConnectionLostHandler(func(client mqtt.Client, err error) {
            log.Printf("Disconnected: %s", err)
        })

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (t *MQTT) onConnect(client mqtt.Client) {
	log.Println("Connected")

    // Subscribe to topics
    subscriptions := map[string]mqtt.MessageHandler{
        fmt.Sprintf("%s/255/255/3/0/3", t.Settings.SubTopic): t.idRequest,
        fmt.Sprintf("%s/+/255/4/0/0", t.Settings.SubTopic):   t.configurationRequest,
        fmt.Sprintf("%s/+/255/4/0/2", t.Settings.SubTopic):   t.dataRequest,
        "mysensors/bootloader/+/+":                           t.bootloaderCommand,
    }

    for topic, handler := range subscriptions {
        if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
            log.Print(token.Error())
        }
    }
}

func (t *MQTT) idRequest(client mqtt.Client, msg mqtt.Message) {
	t.publish(client, fmt.Sprintf("%s/255/255/3/0/4", t.Settings.PubTopic), t.Control.IDRequest())
}

func (t *MQTT) configurationRequest(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	// Attempt to run any bootloader commands
	if t.runBootloaderCommand(client, to) {
		return
	}

	t.publish(client, fmt.Sprintf("%s/%s/255/4/0/1", t.Settings.PubTopic, to), t.Control.ConfigurationRequest(to, payload))
}

// DataRequest - Handle incoming firmware requests
func (t *MQTT) dataRequest(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	t.publish(client, fmt.Sprintf("%s/%s/255/4/0/3", t.Settings.PubTopic, to), t.Control.DataRequest(to, payload))
}

func (t *MQTT) bootloaderCommand(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())

	parts := strings.Split(topic, "/")
	to := parts[2]
	cmd := parts[3]

	t.Control.BootloaderCommand(to, cmd, payload)
}

func (t *MQTT) runBootloaderCommand(client mqtt.Client, to string) bool {
	if blcmd, ok := t.Control.Commands[to]; ok {
		outTopic := fmt.Sprintf("%s/%s/255/4/0/1", t.Settings.PubTopic, to)
		outPayload := blcmd.String()
		t.publish(client, outTopic, outPayload)

		delete(t.Control.Commands, to)
		return true
	}

	return false
}

func (t *MQTT) publish(client mqtt.Client, topic string, payload string) {
	client.Publish(topic, 0, false, payload)
	t.LastPublished = fmt.Sprintf("%s %s", topic, payload)
}
