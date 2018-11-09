package main

import (
	"fmt"
	"os"

	"github.com/kierdavis/ihex-go"
)

// firmware - The MySysBootloader firmware representation
type firmware struct {
	Blocks uint16
	Crc    uint16
	data   []byte
}

// newFirmware - Loads a filename; computes block count and crc
func newFirmware(filename string) *firmware {
	file, err := os.Open(filename)
	if err != nil {
		return &firmware{}
	}
	defer file.Close()

	blocks := uint16(0)
	crc := uint16(0)
	data := []byte{}
	start := uint16(0)
	end := uint16(0)

	scanner := ihex.NewDecoder(file)
	for scanner.Scan() {
		record := scanner.Record()
		if record.Type != ihex.Data {
			continue
		}

		if start == 0 && end == 0 {
			start = record.Address
			end = record.Address
		}

		for record.Address > end {
			data = append(data, 255)
			end++
		}

		data = append(data, record.Data...)

		end += uint16(len(record.Data))
	}

	pad := end % 128
	for i := uint16(0); i < 128-pad; i++ {
		data = append(data, 255)
		end++
	}

	blocks = uint16(end-start) / firmwareBlockSize
	crc = 0xFFFF
	for i := 0; i < len(data); i++ {
		crc = (crc ^ uint16(data[i]&0xFF))
		for j := 0; j < 8; j++ {
			a001 := (crc & 1) > 0
			crc = (crc >> 1)
			if a001 {
				crc = (crc ^ 0xA001)
			}
		}
	}

	return &firmware{
		Blocks: blocks,
		Crc:    crc,
		data:   data,
	}
}

// dataForBlock - Gets a specific block from the firmware data
func (f firmware) dataForBlock(block uint16) ([]byte, error) {
	fromBlock := block * firmwareBlockSize
	toBlock := fromBlock + firmwareBlockSize
	if dataLen := uint16(len(f.data)); dataLen < toBlock {
		return []byte{}, fmt.Errorf("Block %d cannot be found in the firmware data", block)
	}

	data := f.data[fromBlock:toBlock]
	return append(make([]byte, 0, len(data)), data...), nil
}
