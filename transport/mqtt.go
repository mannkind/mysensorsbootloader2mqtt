package transport

import (
	"fmt"
	"log"
	"strings"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/mannkind/mysb/ota"
)

// MQTT - MQTT all the things!
type MQTT struct {
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
func (t *MQTT) Start() error {
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

func (t *MQTT) onConnect(client mqtt.Client) {
	log.Println("Connected to MQTT")

	// Subscribe to topics
	subscriptions := map[string]mqtt.MessageHandler{
		fmt.Sprintf("%s/255/255/3/0/3", t.Settings.SubTopic): t.idRequest,
		fmt.Sprintf("%s/+/255/3/0/22", t.Settings.SubTopic):  t.heartbeatResponse,
		fmt.Sprintf("%s/+/255/4/0/0", t.Settings.SubTopic):   t.configurationRequest,
		fmt.Sprintf("%s/+/255/4/0/2", t.Settings.SubTopic):   t.dataRequest,
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

func (t *MQTT) idRequest(client mqtt.Client, msg mqtt.Message) {
	if t.Control.AutoIDEnabled {
		t.publish(client, fmt.Sprintf("%s/255/255/3/0/4", t.Settings.PubTopic), t.Control.IDRequest())
	}
}

func (t *MQTT) heartbeatResponse(client mqtt.Client, msg mqtt.Message) {
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

func (t *MQTT) configurationRequest(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	t.publish(client, fmt.Sprintf("%s/%s/255/4/0/1", t.Settings.PubTopic, to), t.Control.ConfigurationRequest(to, payload))
}

// DataRequest - Handle incoming firmware requests
func (t *MQTT) dataRequest(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	t.publish(client, fmt.Sprintf("%s/%s/255/4/0/3", t.Settings.PubTopic, to), t.Control.DataRequest(to, payload))
}

func (t *MQTT) queuedCommand(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	t.Control.QueuedCommand(to, topic, payload)
}

func (t *MQTT) publish(client mqtt.Client, topic string, payload string) {
	if token := client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
		log.Printf("Publish Error: %s", token.Error())
	}
	t.LastPublished = fmt.Sprintf("%s %s", topic, payload)
}
