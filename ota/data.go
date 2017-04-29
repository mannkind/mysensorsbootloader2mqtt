package ota

// Data - The MySysBootloader Firmware Config Request
type Data struct {
	Type    uint16
	Version uint16
	Block   uint16
}

// NewData - Loads a string; computes type/version/blocks/crc
func NewData(payload string) *Data {
	t := Data{}
	decodeHexIntoStruct(payload, &t)

	return &t
}

func (t *Data) String(input []byte) string {
	return encodeStructIntoHex(t, input)
}
