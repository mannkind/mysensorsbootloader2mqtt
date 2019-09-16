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

type mqttClient struct {
	*twomqtt.MQTTProxy
	mqttClientConfig
}

func newMQTTClient(mqttClientCfg mqttClientConfig, client *twomqtt.MQTTProxy) *mqttClient {
	c := mqttClient{
		MQTTProxy:        client,
		mqttClientConfig: mqttClientCfg,
	}

	c.Initialize(
		c.onConnect,
		c.onDisconnect,
	)

	return &c
}

func (t *mqttClient) run() {
	t.Run()
}

func (t *mqttClient) onConnect(client mqtt.Client) {
	log.Info("Connected to MQTT")

	// Subscribe to topics
	subscriptions := map[string]mqtt.MessageHandler{
		fmt.Sprintf(idRequestTopic, t.SubTopic):             t.idRequest,
		fmt.Sprintf(firmwareConfigRequestTopic, t.SubTopic): t.configurationRequest,
		fmt.Sprintf(firmwareRequestTopic, t.SubTopic):       t.dataRequest,
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
	if !t.AutoIDEnabled {
		return
	}

	t.NextID++

	log.WithFields(log.Fields{
		"id": t.NextID,
	}).Info("Assigning ID")

	t.publish(fmt.Sprintf(idResponseTopic, t.PubTopic), fmt.Sprintf("%d", t.NextID))
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

	respTopic := fmt.Sprintf(firmwareConfigResponseTopic, t.PubTopic, to)
	respPayload := resp.String()

	log.WithFields(log.Fields{
		"to":   to,
		"req":  req.String(),
		"resp": respPayload,
	}).Info("Configuration Request")
	t.publish(respTopic, respPayload)
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
	respTopic := fmt.Sprintf(firmwareResponseTopic, t.PubTopic, to)
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

	t.publish(respTopic, respPayload)
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
	if t.BootloaderCommands == nil {
		t.BootloaderCommands = make(bootloaderCmdMap)
	}
	t.BootloaderCommands[to] = resp
}

func (t *mqttClient) runBootloaderCommand(client mqtt.Client, to string) bool {
	if blcmd, ok := t.BootloaderCommands[to]; ok {
		outTopic := fmt.Sprintf(firmwareConfigResponseTopic, t.PubTopic, to)
		outPayload := blcmd.String()
		t.publish(outTopic, outPayload)

		delete(t.BootloaderCommands, to)
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
		fw.Path = fmt.Sprintf("%s/%d/%d/firmware.hex", t.FirmwareBasePath, fw.Type, fw.Version)
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
	nodeSettings := t.Nodes[nodeID]
	fw.Type = nodeSettings.Type
	fw.Version = nodeSettings.Version
	fw.Path = fmt.Sprintf("%s/%d/%d/firmware.hex", t.FirmwareBasePath, fw.Type, fw.Version)
	fw.Source = source

	return fw
}

func (t *mqttClient) msgParts(msg mqtt.Message) (string, string, string) {
	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	return topic, payload, to
}

func (t *mqttClient) publish(topic string, payload string) {
	t.PublishWithOpts(topic, payload, twomqtt.MQTTProxyPublishOptions{
		Retained:       false,
		DuplicateCheck: false,
	})
}
