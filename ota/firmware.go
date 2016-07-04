package ota

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

const firmwareBlockSize uint16 = 16

// Firmware - The MySysBootloader Firmware Calculations
type Firmware struct {
	Blocks uint16
	Crc    uint16
	data   []byte
}

// Load - Loads a filename; computes block count and crc
func (f *Firmware) Load(filename string) error {
	file, oErr := os.Open(filename)
	if oErr != nil {
		return oErr
	}

	defer file.Close()

	// create a new scanner and read the file line by line
	scanner := bufio.NewScanner(file)
	var fwdata []byte
	start := uint16(0)
	end := uint16(0)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		for line[0] != ':' {
			line = line[1:]
		}

		rlen := f.parseUint16(line[1:3])
		offset := f.parseUint16(line[3:7])
		rtype := f.parseUint16(line[7:9])

		data := line[9 : 9+(2*rlen)]

		if rtype != 0 {
			continue
		}

		if start == 0 && end == 0 {
			start = offset
			end = offset
		}

		for offset > end {
			fwdata = append(fwdata, 255)
			end++
		}

		for i := uint16(0); i < rlen; i++ {
			double := i * 2
			d := f.parseUint16(data[double : double+2])
			fwdata = append(fwdata, byte(d))
		}

		end += rlen
	}

	pad := end % 128
	for i := uint16(0); i < 128-pad; i++ {
		fwdata = append(fwdata, 255)
		end++
	}

	blocks := (end - start) / firmwareBlockSize
	crc := uint16(0xFFFF)
	for i := uint16(0); i < blocks*firmwareBlockSize; i++ {
		crc = (crc ^ uint16(fwdata[i]&0xFF))
		for j := 0; j < 8; j++ {
			a001 := (crc & 1) > 0
			crc = (crc >> 1)
			if a001 {
				crc = (crc ^ 0xA001)
			}
		}
	}

	f.Blocks = blocks
	f.Crc = crc
	f.data = fwdata

	return scanner.Err()
}

// Data - Gets a specific block from the firmware data
func (f Firmware) Data(block uint16) ([]byte, error) {
	fromBlock := block * firmwareBlockSize
	toBlock := fromBlock + firmwareBlockSize
	if dataLen := uint16(len(f.data)); dataLen < toBlock {
		return []byte{}, fmt.Errorf("Block %d cannot be found in the firmware data.", block)
	}

	return f.data[fromBlock:toBlock], nil
}

func (f Firmware) parseUint16(input string) uint16 {
	if val, err := strconv.ParseUint(input, 16, 16); err == nil {
		return uint16(val)
	}

	return 0
}
