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

// NewData - Loads a string; computes type/version/blocks/crc
func NewData(payload string) *Data {
	t := Data{}
	b, err := hex.DecodeString(payload)
	if err != nil {
		return &t
	}

	r := bytes.NewReader(b)
	binary.Read(r, binary.LittleEndian, &t)

	return &t
}

func (t *Data) String(input []byte) string {
	w := new(bytes.Buffer)
	binary.Write(w, binary.LittleEndian, t)
	return strings.ToUpper(
		strings.Join(
			[]string{hex.EncodeToString(w.Bytes()), hex.EncodeToString(input)},
			"",
		),
	)
}
