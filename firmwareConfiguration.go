package main

// firmwareConfiguration - The MySysBootloader firmware configuration
type firmwareConfiguration struct {
	Type    uint16
	Version uint16
	Blocks  uint16
	Crc     uint16
}

// newFirmwareConfiguration - Loads a string; computes type/version/blocks/crc
func newFirmwareConfiguration(payload string) *firmwareConfiguration {
	t := firmwareConfiguration{}
	hex2Struct(payload, &t)

	return &t
}

func (t *firmwareConfiguration) String() string {
	return struct2Hex(t, nil)
}
