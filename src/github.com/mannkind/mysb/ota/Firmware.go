package ota

import (
	"bufio"
	"os"
	"strconv"
)

const firmwareBlockSize uint16 = 16

// Firmware - The MySysBootloader Firmware Calculations
type Firmware struct {
	Blocks uint16
	Crc    uint16
	Data   []byte
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

		rlen, _ := f.parseUint16(line[1:3])
		offset, _ := f.parseUint16(line[3:7])
		rtype, _ := f.parseUint16(line[7:9])

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
			d, _ := f.parseUint16(data[double : double+2])
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
			if (crc & 1) > 0 {
				crc = ((crc >> 1) ^ 0xA001)
			} else {
				crc = (crc >> 1)
			}
		}
	}

	f.Blocks = blocks
	f.Crc = crc
	f.Data = fwdata

	return scanner.Err()
}

// GetBlock - Gets a specific block from the firmware data
func (f Firmware) GetBlock(block uint16) []byte {
	fromBlock := block * firmwareBlockSize
	toBlock := fromBlock + firmwareBlockSize
	return f.Data[fromBlock:toBlock]
}

func (f Firmware) parseUint16(input string) (uint16, error) {
	val, err := strconv.ParseUint(input, 16, 16)
	if err != nil {
		return uint16(0), err
	}

	return uint16(val), nil
}
