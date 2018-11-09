package main

// firmwareSource - The source of firmware
type firmwareSource int

// firmwareInformation - Structured information about the firmware
type firmwareInformation struct {
	Type    uint16
	Version uint16
	Path    string
	Source  firmwareSource
}

// firmwareSource - The source of firmware
const (
	fwUnknown firmwareSource = iota
	fwNode
	fwReq
	fwDefault
)
