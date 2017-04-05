package controller

import (
	"fmt"
	"log"
	"strings"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/mannkind/mysb/ota"
)

// Go-ified MySensors Constants
const (
	broadcastAddress         = 255
	cInternal                = 3
	cStream                  = 4
	iIDRequest               = 3
	iIDResponse              = 4
	iPreSleepNotification    = 32
	stFirmwareConfigRequest  = 0
	stFirmwareConfigResponse = 1
	stFirmwareRequest        = 2
	stFirmwareResponse       = 3
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
	Control       ota.Control
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
	if token := t.Client.Connect(); token.Wait() && token.Error() != nil {
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
		fmt.Sprintf("%s/%d/%d/%d/0/%d", t.Settings.SubTopic, broadcastAddress, broadcastAddress, cInternal, iIDRequest): t.idRequest,
		fmt.Sprintf("%s/+/%d/%d/0/%d", t.Settings.SubTopic, broadcastAddress, cInternal, iPreSleepNotification):         t.presleepResponse,
		fmt.Sprintf("%s/+/%d/%d/0/%d", t.Settings.SubTopic, broadcastAddress, cStream, stFirmwareConfigRequest):         t.configurationRequest,
		fmt.Sprintf("%s/+/%d/%d/0/%d", t.Settings.SubTopic, broadcastAddress, cStream, stFirmwareRequest):               t.dataRequest,
		"mysensors/bootloader/+/+":                                                                                      t.bootloaderCommand,
	}

	// Subscribe to battery node messages for queuing purposes
	for node, settings := range t.Control.Nodes {
		if settings.QueueMessages {
			sub := fmt.Sprintf("%s/%s/#", t.Settings.PubTopic, node)
			subscriptions[sub] = t.queuedCommand
		}
	}

	//
	if !client.IsConnected() {
		log.Print("Subscribe Error: Not Connected (Reloading Config?)")
		return
	}

	for topic, handler := range subscriptions {
		if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
			log.Printf("Subscribe Error: %s", token.Error())
		}
	}
}

func (t *MysbMQTT) idRequest(client mqtt.Client, msg mqtt.Message) {
	if t.Control.AutoIDEnabled {
		t.publish(client, fmt.Sprintf("%s/%d/%d/%d/0/%d", t.Settings.PubTopic, broadcastAddress, broadcastAddress, cInternal, iIDResponse), t.Control.IDRequest())
	}
}

func (t *MysbMQTT) presleepResponse(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	to := strings.Split(topic, "/")[1]

	if t.Control.Commands[to] == nil || len(t.Control.Commands[to]) == 0 {
		return
	}

	// Unsubscribe; Republish all the commands; Resubscribe
	sub := fmt.Sprintf("%s/%s/#", t.Settings.PubTopic, to)

	if unsubtoken := client.Unsubscribe(sub); unsubtoken.Wait() && unsubtoken.Error() != nil {
		log.Printf("Unsubscribe Error: %s", unsubtoken.Error())
	}

	for _, cmd := range t.Control.Commands[to] {
		log.Printf("Queued Command (Republish): To: %s; Topic: %s; Payload: %s\n", to, cmd.Topic, cmd.Payload)
		t.publish(client, cmd.Topic, cmd.Payload)
	}
	t.Control.QueuedCommand(to, "", "")

	if subtoken := client.Subscribe(sub, 0, t.queuedCommand); subtoken.Wait() && subtoken.Error() != nil {
		log.Printf("Subscribe Error: %s", subtoken.Error())
	}
}

func (t *MysbMQTT) configurationRequest(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	// Attempt to run any bootloader commands
	if t.runBootloaderCommand(client, to) {
		return
	}

	t.publish(
		client,
		fmt.Sprintf("%s/%s/%d/%d/0/%d", t.Settings.PubTopic, to, broadcastAddress, cStream, stFirmwareConfigResponse),
		t.Control.ConfigurationRequest(to, payload),
	)
}

func (t *MysbMQTT) dataRequest(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	t.publish(
		client,
		fmt.Sprintf("%s/%s/%d/%d/0/%d", t.Settings.PubTopic, to, broadcastAddress, cStream, stFirmwareResponse),
		t.Control.DataRequest(to, payload),
	)
}

func (t *MysbMQTT) bootloaderCommand(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())

	parts := strings.Split(topic, "/")
	to := parts[2]
	cmd := parts[3]

	t.Control.BootloaderCommand(to, cmd, payload)
}

func (t *MysbMQTT) runBootloaderCommand(client mqtt.Client, to string) bool {
	if blcmd, ok := t.Control.BootloaderCommands[to]; ok {
		outTopic := fmt.Sprintf("%s/%s/%d/%d/0/%d", t.Settings.PubTopic, to, broadcastAddress, cStream, stFirmwareConfigResponse)
		outPayload := blcmd.String()
		t.publish(client, outTopic, outPayload)

		delete(t.Control.BootloaderCommands, to)
		return true
	}

	return false
}

func (t *MysbMQTT) queuedCommand(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	t.Control.QueuedCommand(to, topic, payload)
}

func (t *MysbMQTT) publish(client mqtt.Client, topic string, payload string) {
	if token := client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
		log.Printf("Publish Error: %s", token.Error())
	}
	t.LastPublished = fmt.Sprintf("%s %s", topic, payload)
}
