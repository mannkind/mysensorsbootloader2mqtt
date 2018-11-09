package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strings"
)

func hex2Struct(s string, data interface{}) error {
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)
	binary.Read(r, binary.LittleEndian, data)

	return nil
}

func struct2Hex(data interface{}, input []byte) string {
	w := new(bytes.Buffer)
	binary.Write(w, binary.LittleEndian, data)
	return strings.ToUpper(
		strings.Join(
			[]string{hex.EncodeToString(w.Bytes()), hex.EncodeToString(input)},
			"",
		),
	)
}
