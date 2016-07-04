package ota

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// Control - Control the interaction of Transport and OTA
type Control struct {
	NextID           uint8
	FirmwareBasePath string
	Nodes            map[string]map[string]string
	Commands         map[string]Configuration
}

// fwSource - The source of firmware
type fwSource int

// fwInfo - Structured information about the firmware
type fwInfo struct {
	Type    string
	Version string
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

// firmwareInfo - Get the firmware to use given a nodeid, or firmware type/version
func (c Control) firmwareInfo(nodeID string, firmwareType string, firmwareVersion string) fwInfo {
	fw := fwInfo{
		Type:    "0",
		Version: "0",
		Path:    "",
		Source:  fwUnknown,
	}

	// Attempt to load firmware from the assignment in config.yaml
	nodeMapping := c.Nodes[nodeID]
	if nodeMapping != nil {
		fw.Type, _ = nodeMapping["type"]
		fw.Version, _ = nodeMapping["version"]
		fw.Path = fmt.Sprintf("%s/%s/%s/firmware.hex", c.FirmwareBasePath, fw.Type, fw.Version)
		fw.Source = fwNode
	}

	// Attempt to load firmware based on the node's request
	if _, err := os.Stat(fw.Path); err != nil {
		fw.Type, fw.Version, fw.Source = firmwareType, firmwareVersion, fwReq
		fw.Path = fmt.Sprintf("%s/%s/%s/firmware.hex", c.FirmwareBasePath, fw.Type, fw.Version)
	}

	// Attempt to laod the default firmware
	if _, err := os.Stat(fw.Path); err != nil {
		defaultMapping := c.Nodes["default"]
		if defaultMapping != nil {
			fw.Type, _ = defaultMapping["type"]
			fw.Version, _ = defaultMapping["version"]
			fw.Source = fwDefault
		}
		fw.Path = fmt.Sprintf("%s/%s/%s/firmware.hex", c.FirmwareBasePath, fw.Type, fw.Version)
	}

	// Awww, nothing worked...
	if _, err := os.Stat(fw.Path); err != nil {
		fw.Type, fw.Version, fw.Path, fw.Source = "0", "0", "", fwUnknown
	}

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
	req := Configuration{}
	req.Load(payload)

	fw := c.firmwareInfo(to, fmt.Sprintf("%d", req.Type), fmt.Sprintf("%d", req.Version))

	firmware := Firmware{}
	firmware.Load(fw.Path)

	resp := Configuration{
		Type:    c.parseUint16(fw.Type, 16),
		Version: c.parseUint16(fw.Version, 16),
		Blocks:  firmware.Blocks,
		Crc:     firmware.Crc,
	}

	log.Printf("Configuration Request: From: %s; Request: %s; Response: %s\n", to, req.String(), resp.String())
	return resp.String()
}

// DataRequest - Handle incoming firmware requests
func (c *Control) DataRequest(to string, payload string) string {
	req := Data{}
	req.Load(payload)

	fw := c.firmwareInfo(to, fmt.Sprintf("%d", req.Type), fmt.Sprintf("%d", req.Version))

	firmware := Firmware{}
	firmware.Load(fw.Path)

	resp := Data{
		Type:    c.parseUint16(fw.Type, 16),
		Version: c.parseUint16(fw.Version, 16),
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

	data, err := firmware.Data(req.Block)
	if err != nil {
		log.Panic(err)
	}

	return resp.String(data)
}

// BootloaderCommand - Handle bootloader commands
// Bootloader commands:
// * 0x01 - Erase EEPROM
// * 0x02 - Set NodeID
// * 0x03 - Set ParentID
func (c *Control) BootloaderCommand(to string, cmd string, payload string) {
	resp := Configuration{
		Type:    c.parseUint16(cmd, 10),
		Version: 0,
		Blocks:  0,
		Crc:     0xDA7A,
	}

	if resp.Type == 0x02 || resp.Type == 0x03 {
		resp.Version = c.parseUint16(payload, 10)
	}

	log.Printf("Bootloader Command: To: %s; Cmd: %s; Payload: %s\n", to, cmd, payload)
	if c.Commands == nil {
		c.Commands = make(map[string]Configuration)
	}
	c.Commands[to] = resp
}

func (c Control) parseUint16(input string, base int) uint16 {
	if val, err := strconv.ParseUint(input, base, 16); err == nil {
		return uint16(val)
	}

	return 0
}
