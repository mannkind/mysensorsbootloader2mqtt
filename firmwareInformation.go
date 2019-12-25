package main

// firmwareInformation - Structured information about the firmware
type firmwareInformation struct {
	Type    uint16
	Version uint16
	Path    string
	Source  firmwareSource
}
