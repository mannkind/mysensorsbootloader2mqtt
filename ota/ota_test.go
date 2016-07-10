package ota

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

var firmwareTests = []struct {
	File    string
	Encoded string
	Blocks  uint16
	Crc     uint16
}{
	{"../test_files/firmware.hex", "../test_files/firmware.encoded", 80, 18132},
	{"../test_files/firmware2.hex", "../test_files/firmware2.encoded", 1304, 19151},
	{"../test_files/firmware3.hex", "../test_files/firmware3.encoded", 1072, 64648},
}

func TestLoadFirmware(t *testing.T) {
	for _, f := range firmwareTests {
		firmware := NewFirmware(f.File)

		if firmware.Blocks != f.Blocks {
			t.Errorf("Incorrect Blocks: Actual: %d; Expected %d", firmware.Blocks, f.Blocks)
		}

		if firmware.Crc != f.Crc {
			t.Errorf("Incorrect Crc: Actual: %d; Expected %d", firmware.Crc, f.Crc)
		}
	}
}

func TestLoadFirmwareGetBlock(t *testing.T) {
	for _, f := range firmwareTests {
		firmware := NewFirmware(f.File)

		if _, err := firmware.Data(f.Blocks); err == nil {
			t.Errorf("Requested a block %d that should not have existed; this should have errored.", f.Blocks)
		}
	}
}

func TestBadConfigurationRequest(t *testing.T) {
	if c := NewConfiguration("0Z00000000000000"); c == nil {
		t.Error("Z is not a valid hexidecmial character and should have errored.")
	}
}

func TestConfigurationRequest(t *testing.T) {
	var tests = []struct {
		Hex     string
		Type    uint16
		Version uint16
		Blocks  uint16
		Crc     uint16
	}{
		{"0000000000000000", 0, 0, 0, 0},
		{"010004005000D446", 1, 4, 80, 18132},
		{"020002003D016A2C", 2, 2, 317, 11370},
		{"0B0001001100F329", 11, 1, 17, 10739},
	}

	for _, v := range tests {
		c := NewConfiguration(v.Hex)

		if c.Type != v.Type {
			t.Errorf("Type does not match. Actual: %d. Expected %d.", c.Type, v.Type)
		}

		if c.Version != v.Version {
			t.Errorf("Version does not match. Actual: %d. Expected %d.", c.Version, v.Version)
		}

		if c.Blocks != v.Blocks {
			t.Errorf("Blocks does not match. Actual: %d. Expected %d.", c.Blocks, v.Blocks)
		}

		if c.Crc != v.Crc {
			t.Errorf("Crc does not match. Actual: %d. Expected %d.", c.Crc, v.Crc)
		}

		if c.String() != v.Hex {
			t.Errorf("Hex does not match. Actual: %s. Expected %s.", c.String(), v.Hex)
		}
	}
}

func TestBadDataRequest(t *testing.T) {
	if d := NewData("0Z00000000000000"); d.Block != 0 || d.Type != 0 || d.Version != 0 {
		t.Error("Z is not a valid hexidecmial character and should have errored.")
	}
}

func TestDataRequest(t *testing.T) {
	for _, v := range firmwareTests {
		firmware := NewFirmware(v.File)
		fmt.Printf("Testing %s\n", v.File)

		file, err := os.Open(v.Encoded)
		if err != nil {
			t.Errorf("Unable to open %s", v.Encoded)
		}
		defer file.Close()

		var payloads []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			payloads = append(payloads, scanner.Text())
		}

		for i := uint16(0); i < v.Blocks; i++ {
			block := v.Blocks - i - 1
			blockHex := fmt.Sprintf("%X", block)
			if len(blockHex) == 1 {
				blockHex = "0" + blockHex
			}
			incoming := strings.Join([]string{"01000100", blockHex, "00"}, "")

			r := NewData(incoming)
			expected := payloads[i]
			data, err := firmware.Data(block)
			if err != nil {
				t.Error(err)
			}

			if actual := r.String(data); actual != expected {
				t.Errorf("Payload does not match. Actual: %s. Expected: %s.", actual, expected)
			}
		}
		fmt.Printf("Done Testing %s\n", v.File)
	}
}

func TestNoFileFirmware(t *testing.T) {
	if firmware := NewFirmware("/tmp/AFileThatDoesNotExist.hex"); firmware.Blocks != 0 || firmware.Crc != 0 {
		t.Error("The file does not exist should have errored.")
	}
}

func defaultTestControl() *Control {
	control := Control{
		NextID:           12,
		FirmwareBasePath: "../test_files",
		Nodes: map[string]map[string]string{
			"default": {
				"type": "1", "version": "1",
			},
			"1": {
				"type": "1", "version": "1",
			},
			"2": {
				"type": "11", "version": "1",
			},
		},
	}
	return &control
}

func TestControlIDRequest(t *testing.T) {
	myControl := defaultTestControl()

	expected := "13"
	if actual := myControl.IDRequest(); actual != expected {
		t.Errorf("Wrong payload - Actual: %s Expected: %s\n", actual, expected)
	}
}

func TestControlConfigurationRequest(t *testing.T) {
	myControl := defaultTestControl()

	expected := "010001005000D446"
	if actual := myControl.ConfigurationRequest("1", "010001005000D446"); actual != expected {
		t.Errorf("Wrong payload - Actual: %s Expected: %s\n", actual, expected)
	}
}

func TestControlDataRequest(t *testing.T) {
	myControl := defaultTestControl()

	fwTest := firmwareTests[0]
	file, err := os.Open(fwTest.Encoded)
	if err != nil {
		t.Errorf("Unable to open %s", fwTest.Encoded)
	}
	defer file.Close()

	var payloads []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		payloads = append(payloads, scanner.Text())
	}

	for i := uint16(0); i < fwTest.Blocks; i++ {
		block := fwTest.Blocks - i - 1
		blockHex := fmt.Sprintf("%X", block)
		if len(blockHex) == 1 {
			blockHex = "0" + blockHex
		}
		request := strings.Join([]string{"01000100", blockHex, "00"}, "")

		expected := payloads[i]
		if actual := myControl.DataRequest("1", request); actual != expected {
			t.Errorf("Payload does not match. Actual: %s. Expected: %s.", actual, expected)
		}
	}
}

func TestControlFirmwareInfoByNode(t *testing.T) {
	var tests = []struct {
		Node       string
		ReqType    uint16
		ReqVersion uint16
		Type       uint16
		Version    uint16
		Path       string
		Source     fwSource
	}{
		{"1", 5, 1, 1, 1, "../test_files/1/1/firmware.hex", fwNode},
		{"2", 2, 1, 11, 1, "../test_files/11/1/firmware.hex", fwNode},
		{"254", 1, 1, 1, 1, "../test_files/1/1/firmware.hex", fwReq},
		{"254", 254, 254, 1, 1, "../test_files/1/1/firmware.hex", fwDefault},
		{"254", 254, 254, 0, 0, "", fwUnknown},
	}

	myControl := defaultTestControl()

	for _, v := range tests {
		if v.Source == fwUnknown {
			delete(myControl.Nodes, "default")
		}

		if fmInfo := myControl.firmwareInfo(v.Node, v.ReqType, v.ReqVersion); fmInfo.Type != v.Type || fmInfo.Version != v.Version || fmInfo.Path != v.Path || fmInfo.Source != v.Source {
			t.Errorf("Unexpected type/version/filename - Actual: %d, %d, %s; Expected: %d, %d, %s", fmInfo.Type, fmInfo.Version, fmInfo.Path, v.Type, v.Version, v.Path)
		}
	}
}

func TestControlBootloaderCmd(t *testing.T) {
	var tests = []struct {
		To      string
		Cmd     string
		Payload string
		Type    uint16
		Version uint16
	}{
		{"1", "1", "", 1, 0},
		{"2", "2", "13", 2, 13},
	}

	for _, v := range tests {
		myControl := defaultTestControl()
		myControl.BootloaderCommand(v.To, v.Cmd, v.Payload)
		if cmd, ok := myControl.Commands[v.To]; !ok {
			t.Error("Bootloader command not found")
		} else if cmd.Type != v.Type || cmd.Version != v.Version || cmd.Blocks != 0 {
			t.Errorf("Bootloader command (%d, %d) not loaded correctly (%d, %d)", v.Type, v.Version, cmd.Type, cmd.Version)
		}
	}
}
