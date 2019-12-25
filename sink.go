package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mannkind/twomqtt"
	log "github.com/sirupsen/logrus"
)

type sink struct {
	*twomqtt.MQTT
	config sinkOpts
}

func newSink(mqtt *twomqtt.MQTT, config sinkOpts) *sink {
	c := sink{
		MQTT:   mqtt,
		config: config,
	}

	c.MQTT.
		SetSubscribeHandler(c.subscribe).
		Initialize()

	return &c
}

func (t *sink) run() {
	t.Run()
}

func (t *sink) subscribe() {
	// Subscribe to topics
	subscriptions := map[string]mqtt.MessageHandler{
		fmt.Sprintf(idRequestTopic, t.config.SubTopic):             func(client mqtt.Client, msg mqtt.Message) { t.idRequest(client, msg) },
		fmt.Sprintf(firmwareConfigRequestTopic, t.config.SubTopic): func(client mqtt.Client, msg mqtt.Message) { t.configurationRequest(client, msg) },
		fmt.Sprintf(firmwareRequestTopic, t.config.SubTopic):       func(client mqtt.Client, msg mqtt.Message) { t.dataRequest(client, msg) },
		firmwareBootloaderCommandTopic:                             func(client mqtt.Client, msg mqtt.Message) { t.bootloaderCommand(client, msg) },
	}

	for topic, handler := range subscriptions {
		llog := log.WithFields(log.Fields{
			"topic": topic,
		})

		llog.Info("Subscribing to a new topic")
		if token := t.Subscribe(topic, 0, handler); !token.Wait() || token.Error() != nil {
			llog.WithFields(log.Fields{
				"error": token.Error(),
			}).Error("Error subscribing to topic")
		}
	}
}

func (t *sink) idRequest(client mqtt.Client, msg mqtt.Message) twomqtt.MQTTMessage {
	log.Info("ID Request")
	if !t.config.AutoIDEnabled {
		return twomqtt.MQTTMessage{}
	}

	t.config.NextID++

	log.WithFields(log.Fields{
		"id": t.config.NextID,
	}).Info("Assigning ID")

	return t.publish(fmt.Sprintf(idResponseTopic, t.config.PubTopic), fmt.Sprintf("%d", t.config.NextID))
}

func (t *sink) configurationRequest(client mqtt.Client, msg mqtt.Message) twomqtt.MQTTMessage {
	_, payload, to := t.msgParts(msg)

	// Attempt to run any bootloader commands
	if ok, resp := t.runBootloaderCommand(client, to); ok {
		return resp
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

	respTopic := fmt.Sprintf(firmwareConfigResponseTopic, t.config.PubTopic, to)
	respPayload := resp.String()

	log.WithFields(log.Fields{
		"to":   to,
		"req":  req.String(),
		"resp": respPayload,
	}).Info("Configuration Request")

	return t.publish(respTopic, respPayload)
}

func (t *sink) dataRequest(client mqtt.Client, msg mqtt.Message) twomqtt.MQTTMessage {
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
	respTopic := fmt.Sprintf(firmwareResponseTopic, t.config.PubTopic, to)
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

	return t.publish(respTopic, respPayload)
}

// Bootloader commands:
// * 0x01 - Erase EEPROM
// * 0x02 - Set NodeID
// * 0x03 - Set ParentID
func (t *sink) bootloaderCommand(client mqtt.Client, msg mqtt.Message) {
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

	if t.config.BootloaderCommands == nil {
		t.config.BootloaderCommands = make(bootloaderCommandMap)
	}

	t.config.BootloaderCommands[to] = resp
}

func (t *sink) runBootloaderCommand(client mqtt.Client, to string) (bool, twomqtt.MQTTMessage) {
	if blcmd, ok := t.config.BootloaderCommands[to]; ok {
		outTopic := fmt.Sprintf(firmwareConfigResponseTopic, t.config.PubTopic, to)
		outPayload := blcmd.String()

		delete(t.config.BootloaderCommands, to)
		return true, t.publish(outTopic, outPayload)
	}

	return false, twomqtt.MQTTMessage{}
}

func (t *sink) firmwareInfo(nodeID string, firmwareType uint16, firmwareVersion uint16) firmwareInformation {
	fw := firmwareInformation{
		Source: fwUnknown,
	}

	// Attempt to load firmware from the assignment in config.yaml
	fw = t.firmwareInfoAssignment(nodeID, fwNode)

	// Attempt to load firmware based on the node's request
	if _, err := os.Stat(fw.Path); err != nil {
		fw.Type, fw.Version, fw.Source = firmwareType, firmwareVersion, fwReq
		fw.Path = fmt.Sprintf("%s/%d/%d/firmware.hex", t.config.FirmwareBasePath, fw.Type, fw.Version)
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

func (t *sink) firmwareInfoAssignment(nodeID string, source firmwareSource) firmwareInformation {
	fw := firmwareInformation{
		Source: fwUnknown,
	}

	// Attempt to load firmware from the assignment in config.yaml
	nodeSettings := t.config.Nodes[nodeID]
	fw.Type = nodeSettings.Type
	fw.Version = nodeSettings.Version
	fw.Path = fmt.Sprintf("%s/%d/%d/firmware.hex", t.config.FirmwareBasePath, fw.Type, fw.Version)
	fw.Source = source

	return fw
}

func (t *sink) msgParts(msg mqtt.Message) (string, string, string) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	return topic, payload, to
}

func (t *sink) publish(topic string, payload string) twomqtt.MQTTMessage {
	return t.PublishWithOpts(topic, payload, twomqtt.MQTTPublishOpts{
		Retained:       false,
		DuplicateCheck: false,
	})
}
