package controller

import (
	"fmt"
	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/mannkind/mysb/ota"
	"log"
	"strconv"
	"strings"
)

// ControlCfg - Control the interaction of Node and Firmware Uploader
type Control struct {
	NextID   uint8
	SubTopic string
	PubTopic string
	OTA      struct {
		Types    map[string]string
		Versions map[string]string
		Firmware map[string]map[string]string
		Nodes    map[string]map[string]string
	}
	Commands      map[string]ota.Configuration
	LastPublished struct {
		Topic   string
		Payload string
	}
}

// FirmwareInfo - Get the firmware to use given a nodeid, or firmware type/version
func (c Control) FirmwareInfo(nodeID string, firmwareType string, firmwareVersion string) (string, string, string) {
	log.Println("Trying to get the firmware... ")
	outType := "0"
	outVer := "0"
	outPath := ""

	log.Println("\t - assigned to the node")
	nodeMapping := c.OTA.Nodes[nodeID]
	if nodeMapping != nil {
		outType, _ = nodeMapping["type"]
		outVer, _ = nodeMapping["version"]
		outPath = c.OTA.Firmware[outType][outVer]
	}

	if outPath == "" {
		log.Println("\t - by requested type/version")
		outType = firmwareType
		outVer = firmwareVersion
		outPath = c.OTA.Firmware[outType][outVer]
	}

	if outPath == "" {
		log.Println("\t - default firmware")
		outType = "0"
		outVer = "0"
		outPath = c.OTA.Firmware[outType][outVer]
	}

	return outType, outVer, outPath
}

// IDRequest - Handle incoming ID requests
func (c *Control) IDRequest(client *mqtt.Client, msg mqtt.Message) {
	log.Println("IDRequest")

	c.NextID++

	rTopic := fmt.Sprintf("%s/255/255/3/0/4", c.PubTopic)
	rPayload := fmt.Sprintf("%d", c.NextID)

	c.publish(client, rTopic, rPayload)
}

// ConfigurationRequest - Handle incoming firmware configuration requets
func (c *Control) ConfigurationRequest(client *mqtt.Client, msg mqtt.Message) {
	log.Println("ConfigurationRequest")

	topic := msg.Topic()
	//payload := string(msg.Payload())

	to := strings.Split(topic, "/")[1]
	req := ota.Configuration{}

	// Attempt to run any bootloader commands
	if c.runBootloaderCommand(client, to) {
		return
	}

	typ, ver, filename := c.FirmwareInfo(to, fmt.Sprintf("%d", req.Type), fmt.Sprintf("%d", req.Version))
	firmware := ota.Firmware{}
	firmware.Load(filename)

	var ftype uint16
	var fver uint16
	if val, err := strconv.ParseUint(typ, 10, 16); err == nil {
		ftype = uint16(val)
	}
	if val, err := strconv.ParseUint(ver, 10, 16); err == nil {
		fver = uint16(val)
	}
	resp := ota.Configuration{
		Type:    ftype,
		Version: fver,
		Blocks:  firmware.Blocks,
		Crc:     firmware.Crc,
	}
	outTopic := fmt.Sprintf("%s/%s/255/4/0/1", c.PubTopic, to)
	outPayload := resp.String()

	c.publish(client, outTopic, outPayload)
}

// DataRequest - Handle incoming firmware requests
func (c *Control) DataRequest(client *mqtt.Client, msg mqtt.Message) {
	log.Println("DataRequest")

	topic := msg.Topic()
	payload := string(msg.Payload())
	to := strings.Split(topic, "/")[1]

	req := ota.Data{}
	req.Load(payload)

	ftype, fver, fname := c.FirmwareInfo(to, fmt.Sprintf("%d", req.Type), fmt.Sprintf("%d", req.Version))
	firmware := ota.Firmware{}
	firmware.Load(fname)

	if req.Block%50 == 0 {
		log.Printf("Sending block %d of %s %s\n", req.Block, ftype, fver)
	}

	firmwareType, _ := c.parseUint16(ftype)
	firmwareVer, _ := c.parseUint16(fver)
	resp := ota.Data{
		Type:    firmwareType,
		Version: firmwareVer,
		Block:   req.Block,
	}

	outTopic := fmt.Sprintf("%s/%s/255/4/0/3", c.PubTopic, to)
	outPayload := resp.String(firmware.GetBlock(req.Block))

	c.publish(client, outTopic, outPayload)
}

// BootloaderCommand - Handle bootloader commands
func (c *Control) BootloaderCommand(client *mqtt.Client, msg mqtt.Message) {
	log.Println("BootloaderCommand")

	topic := msg.Topic()
	payload := string(msg.Payload())

	parts := strings.Split(topic, "/")
	to := parts[1]
	cmd := parts[2]

	command, _ := c.parseUint16(cmd)
	pl, _ := c.parseUint16(payload)

	resp := ota.Configuration{
		Type:    command,
		Version: 0,
		Blocks:  0,
		Crc:     0xDA7A,
	}

	/*
	 Bootloader commands
	 0x01 - Erase EEPROM
	 0x02 - Set NodeID
	 0x03 - Set ParentID
	*/
	if resp.Type == 0x02 || resp.Type == 0x03 {
		resp.Version = pl
	}

	c.Commands[to] = resp
}

func (c *Control) runBootloaderCommand(client *mqtt.Client, to string) bool {
	if blcmd, ok := c.Commands[to]; ok {
		outTopic := fmt.Sprintf("%s/%s/255/4/0/1", c.PubTopic, to)
		outPayload := blcmd.String()
		c.publish(client, outTopic, outPayload)

		delete(c.Commands, to)
		return true
	}

	return false
}

func (c *Control) publish(client *mqtt.Client, topic string, payload string) {
	client.Publish(topic, 0, false, payload)
	c.LastPublished.Topic = topic
	c.LastPublished.Payload = payload
}

func (c Control) parseUint16(input string) (uint16, error) {
	val, err := strconv.ParseUint(input, 16, 16)
	if err != nil {
		return 0, err
	}

	return uint16(val), nil
}
