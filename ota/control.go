package ota

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Control - Control the interaction of Transport and OTA
type Control struct {
	AutoIDEnabled      bool
	NextID             uint8
	FirmwareBasePath   string
	Nodes              map[string]NodeSettings
	BootloaderCommands map[string]Configuration
	Commands           map[string][]QueuedCommand
}

// NodeSettings - The settings for a node
type NodeSettings struct {
	Type          uint16
	Version       uint16
	QueueMessages bool
}

// QueuedCommand - A queued command for sleeping nodes
type QueuedCommand struct {
	Topic   string
	Payload string
}

// fwSource - The source of firmware
type fwSource int

// fwInfo - Structured information about the firmware
type fwInfo struct {
	Type    uint16
	Version uint16
	Path    string
	Source  fwSource
}

// fwSource - The source of firmware
const (
	fwUnknown fwSource = iota
	fwNode
	fwReq
	fwDefault
)

func (c Control) firmwareInfo(nodeID string, firmwareType uint16, firmwareVersion uint16) fwInfo {
	fw := fwInfo{
		Source: fwUnknown,
	}

	// Attempt to load firmware from the assignment in config.yaml
	fw = c.firmwareInfoAssignment(nodeID, fwNode)

	// Attempt to load firmware based on the node's request
	if _, err := os.Stat(fw.Path); err != nil {
		fw.Type, fw.Version, fw.Source = firmwareType, firmwareVersion, fwReq
		fw.Path = fmt.Sprintf("%s/%d/%d/firmware.hex", c.FirmwareBasePath, fw.Type, fw.Version)
	}

	// Attempt to laod the default firmware
	if _, err := os.Stat(fw.Path); err != nil {
		fw = c.firmwareInfoAssignment("default", fwDefault)
	}

	// Awww, nothing worked...
	if _, err := os.Stat(fw.Path); err != nil {
		fw.Type, fw.Version, fw.Path, fw.Source = 0, 0, "", fwUnknown
	}

	return fw
}

func (c Control) firmwareInfoAssignment(nodeID string, source fwSource) fwInfo {
	fw := fwInfo{
		Source: fwUnknown,
	}

	// Attempt to load firmware from the assignment in config.yaml
	nodeSettings := c.Nodes[nodeID]
	fw.Type = nodeSettings.Type
	fw.Version = nodeSettings.Version
	fw.Path = fmt.Sprintf("%s/%d/%d/firmware.hex", c.FirmwareBasePath, fw.Type, fw.Version)
	fw.Source = source

	return fw
}

// IDRequest - Handle incoming ID requests
func (c *Control) IDRequest() string {
	log.Println("ID Request")
	c.NextID++

	log.Printf("Assigning ID: %d\n", c.NextID)
	return fmt.Sprintf("%d", c.NextID)
}

// ConfigurationRequest - Handle incoming firmware configuration requets
func (c *Control) ConfigurationRequest(to string, payload string) string {
	req := NewConfiguration(payload)
	fw := c.firmwareInfo(to, req.Type, req.Version)
	firmware := NewFirmware(fw.Path)
	resp := Configuration{
		Type:    fw.Type,
		Version: fw.Version,
		Blocks:  firmware.Blocks,
		Crc:     firmware.Crc,
	}

	log.Printf("Configuration Request: From: %s; Request: %s; Response: %s\n", to, req.String(), resp.String())
	return resp.String()
}

// DataRequest - Handle incoming firmware requests
func (c *Control) DataRequest(to string, payload string) string {
	req := NewData(payload)
	fw := c.firmwareInfo(to, req.Type, req.Version)
	firmware := NewFirmware(fw.Path)
	resp := Data{
		Type:    fw.Type,
		Version: fw.Version,
		Block:   req.Block,
	}

	if req.Block+1 == firmware.Blocks {
		log.Printf("Data Request: From: %s; Payload: %s\n", to, payload)
		log.Printf("Sending last block of %d to %s\n", firmware.Blocks, to)
	} else if req.Block == 0 {
		log.Printf("Sending first block of %d to %s\n", firmware.Blocks, to)
	} else if req.Block%50 == 0 {
		log.Printf("Sending block %d of %d to %s\n", req.Block, firmware.Blocks, to)
	}

	data, _ := firmware.Data(req.Block)
	return resp.String(data)
}

// BootloaderCommand - Handle bootloader commands
// Bootloader commands:
// * 0x01 - Erase EEPROM
// * 0x02 - Set NodeID
// * 0x03 - Set ParentID
func (c *Control) BootloaderCommand(to string, cmd string, payload string) {
	blCmd, _ := strconv.ParseUint(cmd, 10, 16)
	resp := Configuration{
		Type:    uint16(blCmd),
		Version: 0,
		Blocks:  0,
		Crc:     0xDA7A,
	}

	if resp.Type == 0x02 || resp.Type == 0x03 {
		blVersion, _ := strconv.ParseUint(payload, 10, 16)
		resp.Version = uint16(blVersion)
	}

	log.Printf("Bootloader Command: To: %s; Cmd: %s; Payload: %s\n", to, cmd, payload)
	if c.BootloaderCommands == nil {
		c.BootloaderCommands = make(map[string]Configuration)
	}
	c.BootloaderCommands[to] = resp
}

// QueuedCommand - Handle queued commands to nodes
func (c *Control) QueuedCommand(to string, topic string, payload string) {
	// Reset any queued commands on blank topic/payload
	if topic == "" && payload == "" {
		log.Printf("Queued Command (Reset): To: %s\n", to)
		if c.Commands == nil {
			c.Commands = make(map[string][]QueuedCommand)
		}
		c.Commands[to] = make([]QueuedCommand, 0)
		return
	}

	parts := strings.Split(topic, "/")
	msgType := parts[3]
	subType := parts[5]

	skipMsgTypes := map[string]bool{
		"0": true, // Presentation
		"4": true, // Stream for OTA
	}

	includeInternalSubTypes := map[string]bool{
		"4":  true, // ID Response
		"8":  true, // ParentID Repsonse
		"13": true, // Reboot
	}

	if skipMsgTypes[msgType] || (msgType == "3" && !includeInternalSubTypes[subType]) {
		return
	}

	if c.Commands == nil {
		c.Commands = make(map[string][]QueuedCommand)
	}

	if c.Commands[to] == nil {
		c.Commands[to] = make([]QueuedCommand, 0)
	}

	log.Printf("Queued Command (Saved): To: %s; Topic: %s; Payload: %s\n", to, topic, payload)
	c.Commands[to] = append(c.Commands[to], QueuedCommand{topic, payload})
}
