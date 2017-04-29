package ota

// Configuration - The MySysBootloader Firmware Config Request
type Configuration struct {
	Type    uint16
	Version uint16
	Blocks  uint16
	Crc     uint16
}

// NewConfiguration - Loads a string; computes type/version/blocks/crc
func NewConfiguration(payload string) *Configuration {
	t := Configuration{}
	decodeHexIntoStruct(payload, &t)

	return &t
}

func (t *Configuration) String() string {
	return encodeStructIntoHex(t, nil)
}
