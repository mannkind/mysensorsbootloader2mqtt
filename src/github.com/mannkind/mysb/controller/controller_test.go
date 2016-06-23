package controller

import (
	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"testing"
)

const testHex = "../test_files/test.hex"

var client = new(mqtt.Client)

func defaultTestControl() Control {
	control := Control{NextID: 12}
	control.OTA.Firmware = make(map[string]map[string]string)
	control.OTA.Firmware["1"] = make(map[string]string)
	control.OTA.Firmware["1"]["1"] = testHex
	control.OTA.Nodes = make(map[string]map[string]string)
	control.OTA.Nodes["1"] = make(map[string]string)
	control.OTA.Nodes["1"]["type"] = "1"
	control.OTA.Nodes["1"]["version"] = "1"

	return control
}

func TestIDRequest(t *testing.T) {
	control := defaultTestControl()
	msg := mockMessage{}
	msg.On("Topic").Return("_/255/255/3/0/3")
	msg.On("Payload").Return([]byte(""))
	control.IDRequest(client, msg)

	if control.NextID != 13 {
		t.Error("NextID didn't increment")
	}

	if control.LastPublished.Topic != "/255/255/3/0/4" {
		t.Error("Published to wrong topic")
	}

	if control.LastPublished.Payload != "13" {
		t.Error("Published to wrong payload")
	}
}

func TestConfigurationRequest(t *testing.T) {
	control := defaultTestControl()

	msg := &mockMessage{}
	msg.On("Topic").Return("_/1/255/4/0/0")
	msg.On("Payload").Return([]byte("010001005000D446"))
	control.ConfigurationRequest(client, msg)
	if control.LastPublished.Topic != "/1/255/4/0/1" {
		t.Error("The topic used was incorrect")
	}

	if control.LastPublished.Payload != "010001005000D446" {
		t.Error("The payload sent was incorrect")
	}
}

func TestDataRequest(t *testing.T) {
	control := defaultTestControl()

	msg := &mockMessage{}
	msg.On("Topic").Return("_/1/255/4/0/0")
	msg.On("Payload").Return([]byte("010001000100"))
	control.DataRequest(client, msg)
	if control.LastPublished.Topic != "/1/255/4/0/3" {
		t.Error("The topic used was incorrect")
	}

	if control.LastPublished.Payload != "0100010001000C946E000C946E000C946E000C946E00" {
		t.Error("The payload sent was incorrect")
	}
}

func TestFirmwareInfoByNode(t *testing.T) {
	control := defaultTestControl()

	if typ, ver, filename := control.FirmwareInfo("1", "1", "1"); typ != "1" && ver != "1" && filename != testHex {
		t.Error("Node: Unexpected node-based type/version/filename")
	}
}

func TestFirmwareInfoByRequest(t *testing.T) {
	control := defaultTestControl()
	control.OTA.Nodes = make(map[string]map[string]string)

	if typ, ver, filename := control.FirmwareInfo("254", "1", "1"); typ != "1" && ver != "1" && filename != testHex {
		t.Error("Requested: Unexpected request-based type/version/filename")
	}
}

func TestFirmwareInfoByDefault(t *testing.T) {
	control := defaultTestControl()
	control.OTA.Firmware = make(map[string]map[string]string)
	control.OTA.Firmware["0"] = make(map[string]string)
	control.OTA.Firmware["0"]["0"] = testHex
	control.OTA.Nodes = make(map[string]map[string]string)
	control.OTA.Nodes["0"] = make(map[string]string)
	control.OTA.Nodes["0"]["type"] = "1"
	control.OTA.Nodes["0"]["version"] = "1"

	if typ, ver, filename := control.FirmwareInfo("254", "254", "254"); typ != "1" && ver != "1" && filename != testHex {
		t.Error("Default: Unexpected default type/version/filename")
	}
}
