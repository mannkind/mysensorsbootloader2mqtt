package main

// firmwareRequest - The MySysBootloader Firmware Config Request
type firmwareRequest struct {
	Type    uint16
	Version uint16
	Block   uint16
}

// newFirmwareRequest - Loads a string; computes type/version/blocks/crc
func newFirmwareRequest(payload string) *firmwareRequest {
	t := firmwareRequest{}
	hex2Struct(payload, &t)

	return &t
}

func (t *firmwareRequest) String(input []byte) string {
	return struct2Hex(t, input)
}
