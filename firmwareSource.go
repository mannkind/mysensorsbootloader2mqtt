package main

// firmwareSource - The source of firmware
type firmwareSource int

// firmwareSource - The source of firmware
const (
	fwUnknown firmwareSource = iota
	fwNode
	fwReq
	fwDefault
)
