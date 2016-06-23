package ota

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strings"
)

// Data - The MySysBootloader Firmware Config Request
type Data struct {
	Type    uint16
	Version uint16
	Block   uint16
}

// Load - Loads a string; computes type/version/blocks/crc
func (t *Data) Load(payload string) error {
	b, err := hex.DecodeString(payload)
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)
	return binary.Read(r, binary.LittleEndian, t)
}

func (t *Data) String(input []byte) string {
	w := new(bytes.Buffer)
	if err := binary.Write(w, binary.LittleEndian, t); err != nil {
		return ""
	}

	base := hex.EncodeToString(w.Bytes())
	data := hex.EncodeToString(input)
	return strings.ToUpper(base + data)
}
