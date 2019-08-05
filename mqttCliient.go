package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttExtDI "github.com/mannkind/paho.mqtt.golang.ext/di"
	log "github.com/sirupsen/logrus"
)

type mqttClient struct {
	subTopic           string
	pubTopic           string
	autoIDEnabled      bool
	nextID             uint
	firmwareBasePath   string
	nodes              nodeSettingsMap
	bootloaderCommands bootloaderCmdMap
	lastPublished      string
	client             mqtt.Client
}

// newMqttClient - Returns a new reference to a fully configured object.
func newMqttClient(config *Config, mqttFuncWrapper *mqttExtDI.MQTTFuncWrapper) *mqttClient {
	m := mqttClient{
		subTopic:         config.SubTopic,
		pubTopic:         config.PubTopic,
		autoIDEnabled:    config.AutoIDEnabled,
		nextID:           config.NextID,
		firmwareBasePath: config.FirmwareBasePath,
		nodes:            config.Nodes,
	}

	opts := mqttFuncWrapper.
		ClientOptsFunc().
		AddBroker(config.MQTT.Broker).
		SetClientID(config.MQTT.ClientID).
		SetOnConnectHandler(m.onConnect).
		SetConnectionLostHandler(m.onDisconnect).
		SetUsername(config.MQTT.Username).
		SetPassword(config.MQTT.Password)

	m.client = mqttFuncWrapper.ClientFunc(opts)

	return &m
}

func (t *mqttClient) run() {
	t.runAfter(0 * time.Second)
}

func (t *mqttClient) runAfter(delay time.Duration) {
	time.Sleep(delay)

	log.Info("Connecting to MQTT")
	if token := t.client.Connect(); !token.Wait() || token.Error() != nil {
		log.WithFields(log.Fields{
			"error": token.Error(),
		}).Error("Error connecting to MQTT")

		delay = t.adjustReconnectDelay(delay)

		log.WithFields(log.Fields{
			"delay": delay,
		}).Info("Sleeping before attempting to reconnect to MQTT")

		t.runAfter(delay)
	}
}

func (t *mqttClient) adjustReconnectDelay(delay time.Duration) time.Duration {
	var maxDelay float64 = 120
	defaultDelay := 2 * time.Second

	// No delay, set to default delay
	if delay.Seconds() == 0 {
		delay = defaultDelay
	} else {
		// Increment the delay
		delay = delay * 2

		// If the delay is above two minutes, reset to default
		if delay.Seconds() > maxDelay {
			delay = defaultDelay
		}
	}

	return delay
}

func (t *mqttClient) onConnect(client mqtt.Client) {
	log.Info("Connected to MQTT")

	// Subscribe to topics
	subscriptions := map[string]mqtt.MessageHandler{
		fmt.Sprintf(idRequestTopic, t.subTopic):             t.idRequest,
		fmt.Sprintf(firmwareConfigRequestTopic, t.subTopic): t.configurationRequest,
		fmt.Sprintf(firmwareRequestTopic, t.subTopic):       t.dataRequest,
		firmwareBootloaderCommandTopic:                      t.bootloaderCommand,
	}

	for topic, handler := range subscriptions {
		llog := log.WithFields(log.Fields{
			"topic": topic,
		})

		llog.Info("Subscribing to a new topic")
		if token := client.Subscribe(topic, 0, handler); !token.Wait() || token.Error() != nil {
			llog.WithFields(log.Fields{
				"error": token.Error(),
			}).Error("Error subscribing to topic")
		}
	}
}

func (t *mqttClient) onDisconnect(client mqtt.Client, err error) {
	log.WithFields(log.Fields{
		"error": err,
	}).Error("Disconnected from MQTT")
}

func (t *mqttClient) idRequest(client mqtt.Client, msg mqtt.Message) {
	log.Info("ID Request")
	if !t.autoIDEnabled {
		return
	}

	t.nextID++

	log.WithFields(log.Fields{
		"id": t.nextID,
	}).Info("Assigning ID")

	t.publish(client, fmt.Sprintf(idResponseTopic, t.pubTopic), fmt.Sprintf("%d", t.nextID))
}

func (t *mqttClient) configurationRequest(client mqtt.Client, msg mqtt.Message) {
	_, payload, to := t.msgParts(msg)

	// Attempt to run any bootloader commands
	if t.runBootloaderCommand(client, to) {
		return
	}

	req := newFirmwareConfiguration(payload)
	fw := t.firmwareInfo(to, req.Type, req.Version)
	firmware := newFirmware(fw.Path)
	resp := firmwareConfiguration{
		Type:    fw.Type,
		Version: fw.Version,
		Blocks:  firmware.Blocks,
		Crc:     firmware.Crc,
	}

	respTopic := fmt.Sprintf(firmwareConfigResponseTopic, t.pubTopic, to)
	respPayload := resp.String()

	log.WithFields(log.Fields{
		"to":   to,
		"req":  req.String(),
		"resp": respPayload,
	}).Info("Configuration Request")
	t.publish(client, respTopic, respPayload)
}

func (t *mqttClient) dataRequest(client mqtt.Client, msg mqtt.Message) {
	_, payload, to := t.msgParts(msg)

	req := newFirmwareRequest(payload)
	fw := t.firmwareInfo(to, req.Type, req.Version)
	firmware := newFirmware(fw.Path)
	resp := firmwareRequest{
		Type:    fw.Type,
		Version: fw.Version,
		Block:   req.Block,
	}

	data, _ := firmware.dataForBlock(req.Block)
	respTopic := fmt.Sprintf(firmwareResponseTopic, t.pubTopic, to)
	respPayload := resp.String(data)

	llog := log.WithFields(log.Fields{
		"reqBlock":   req.Block,
		"totalBlock": firmware.Blocks,
		"to":         to,
		"payload":    payload,
	})
	if req.Block+1 == firmware.Blocks {
		llog.Info("Data Request")
		llog.Info("Sending last block")
	} else if req.Block == 0 {
		llog.Info("Sending first block")
	} else if req.Block%50 == 0 {
		llog.Info("Sending block")
	}

	t.publish(client, respTopic, respPayload)
}

// Bootloader commands:
// * 0x01 - Erase EEPROM
// * 0x02 - Set NodeID
// * 0x03 - Set ParentID
func (t *mqttClient) bootloaderCommand(client mqtt.Client, msg mqtt.Message) {
	topic, payload, _ := t.msgParts(msg)

	parts := strings.Split(topic, "/")
	to := parts[2]
	cmd := parts[3]

	blCmd, _ := strconv.ParseUint(cmd, 10, 16)
	resp := firmwareConfiguration{
		Type:    uint16(blCmd),
		Version: 0,
		Blocks:  0,
		Crc:     0xDA7A,
	}

	if resp.Type == 0x02 || resp.Type == 0x03 {
		blVersion, _ := strconv.ParseUint(payload, 10, 16)
		resp.Version = uint16(blVersion)
	}

	log.WithFields(log.Fields{
		"to":      to,
		"cmd":     cmd,
		"payload": payload,
	}).Info("Bootloader Command")
	if t.bootloaderCommands == nil {
		t.bootloaderCommands = make(bootloaderCmdMap)
	}
	t.bootloaderCommands[to] = resp
}

func (t *mqttClient) runBootloaderCommand(client mqtt.Client, to string) bool {
	if blcmd, ok := t.bootloaderCommands[to]; ok {
		outTopic := fmt.Sprintf(firmwareConfigResponseTopic, t.pubTopic, to)
		outPayload := blcmd.String()
		t.publish(client, outTopic, outPayload)

		delete(t.bootloaderCommands, to)
		return true
	}

	return false
}

func (t *mqttClient) firmwareInfo(nodeID string, firmwareType uint16, firmwareVersion uint16) firmwareInformation {
	fw := firmwareInformation{
		Source: fwUnknown,
	}

	// Attempt to load firmware from the assignment in config.yaml
	fw = t.firmwareInfoAssignment(nodeID, fwNode)

	// Attempt to load firmware based on the node's request
	if _, err := os.Stat(fw.Path); err != nil {
		fw.Type, fw.Version, fw.Source = firmwareType, firmwareVersion, fwReq
		fw.Path = fmt.Sprintf("%s/%d/%d/firmware.hex", t.firmwareBasePath, fw.Type, fw.Version)
	}

	// Attempt to laod the default firmware
	if _, err := os.Stat(fw.Path); err != nil {
		fw = t.firmwareInfoAssignment("default", fwDefault)
	}

	// Awww, nothing worked...
	if _, err := os.Stat(fw.Path); err != nil {
		fw.Type, fw.Version, fw.Path, fw.Source = 0, 0, "", fwUnknown
	}

	return fw
}

func (t *mqttClient) firmwareInfoAssignment(nodeID string, source firmwareSource) firmwareInformation {
	fw := firmwareInformation{
		Source: fwUnknown,
	}

	// Attempt to load firmware from the assignment in config.yaml
	nodeSettings := t.nodes[nodeID]
	fw.Type = nodeSettings.Type
	fw.Version = nodeSettings.Version
	fw.Path = fmt.Sprintf("%s/%d/%d/firmware.hex", t.firmwareBasePath, fw.Type, fw.Version)
	fw.Source = source

	return fw
}

func (t *mqttClient) msgParts(msg mqtt.Message) (string, string, string) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	return topic, payload, to
}

func (t *mqttClient) publish(client mqtt.Client, topic string, payload string) {
	llog := log.WithFields(log.Fields{
		"topic":   topic,
		"payload": payload,
	})

	llog.Info("Publishing to MQTT")
	if token := client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
		log.Error("Publishing error")
	}

	t.lastPublished = fmt.Sprintf("%s %s", topic, payload)
}
