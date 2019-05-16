package main

import (
	"fmt"
	"testing"

	mqttExtCfg "github.com/mannkind/paho.mqtt.golang.ext/cfg"
	mqttExtDI "github.com/mannkind/paho.mqtt.golang.ext/di"
	"gopkg.in/yaml.v2"
)

const nodeRequestHex = "test_files/1/1/firmware.hex"

func defaultTestMQTT() *Mysb {
	var testConfig = `
      nodes:
        default: { type: 1, version: 1 }
        1: { type: 1, version: 1, queueMessages: true }
    `

	mysb := NewMysb(NewConfig(mqttExtCfg.NewMQTTConfig()), mqttExtDI.NewMQTTFuncWrapper())
	mysb.autoIDEnabled = true
	mysb.nextID = 12
	mysb.firmwareBasePath = "test_files"
	if err := yaml.Unmarshal([]byte(testConfig), &mysb); err != nil {
		panic(err)
	}
	return mysb
}

func TestMqttIDRequest(t *testing.T) {
	mysb := defaultTestMQTT()
	var tests = []struct {
		Request       string
		Response      string
		AutoIDEnabled bool
	}{
		{fmt.Sprintf("%s/255/255/3/0/3", mysb.subTopic), fmt.Sprintf("%s/255/255/3/0/4 %s", mysb.pubTopic, "13"), true},
		{fmt.Sprintf("%s/255/255/3/0/3", mysb.subTopic), "", false},
	}
	for _, v := range tests {
		msg := &mockMessage{
			topic:   v.Request,
			payload: []byte(""),
		}

		expected := v.Response

		mysb.lastPublished = ""
		mysb.autoIDEnabled = v.AutoIDEnabled
		mysb.idRequest(mysb.client, msg)
		if mysb.lastPublished != expected {
			t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", mysb.lastPublished, expected)
		}
	}
}

func TestMqttConfigurationRequest(t *testing.T) {
	mysb := defaultTestMQTT()
	msg := &mockMessage{
		topic:   fmt.Sprintf("%s/1/255/4/0/0", mysb.subTopic),
		payload: []byte("010001005000D446"),
	}

	expected := fmt.Sprintf("%s/1/255/4/0/1 %s", mysb.pubTopic, "010001005000D446")
	mysb.configurationRequest(mysb.client, msg)
	if mysb.lastPublished != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", mysb.lastPublished, expected)
	}
}

func TestMqttDataRequest(t *testing.T) {
	mysb := defaultTestMQTT()
	msg := &mockMessage{
		topic:   fmt.Sprintf("%s/1/255/4/0/0", mysb.subTopic),
		payload: []byte("010001000100"),
	}

	expected := fmt.Sprintf("%s/1/255/4/0/3 %s", mysb.pubTopic, "0100010001000C946E000C946E000C946E000C946E00")

	mysb.dataRequest(mysb.client, msg)
	if mysb.lastPublished != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", mysb.lastPublished, expected)
	}
}

func TestMqttBootloaderCommand(t *testing.T) {
	mysb := defaultTestMQTT()
	var tests = []struct {
		To              string
		Cmd             string
		Payload         string
		ExpectedPayload string
	}{
		{"1", "1", "", "0100000000007ADA"},
		{"2", "2", "9", "0200090000007ADA"},
	}

	for _, v := range tests {
		msg := &mockMessage{
			topic:   fmt.Sprintf("mysensors/bootloader/%s/%s", v.To, v.Cmd),
			payload: []byte(v.Payload),
		}

		mysb.bootloaderCommand(mysb.client, msg)
		if _, ok := mysb.bootloaderCommands[v.To]; !ok {
			t.Error("Bootloader command not found")
		} else {
			if ok := mysb.runBootloaderCommand(mysb.client, v.To); !ok {
				t.Error("Bootloader command not run")
			} else {
				expected := fmt.Sprintf("%s/%s/255/4/0/1 %s", mysb.pubTopic, v.To, v.ExpectedPayload)
				if mysb.lastPublished != expected {
					t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", mysb.lastPublished, expected)
				}
			}
		}
	}
}

func TestMqttBadBootloaderCommand(t *testing.T) {
	mysb := defaultTestMQTT()
	if ok := mysb.runBootloaderCommand(mysb.client, "1"); ok {
		t.Error("Bootloader command didn't exist, should not have returned true")
	}
}

func TestMqttRun(t *testing.T) {
	mysb := defaultTestMQTT()
	if err := mysb.Run(); err != nil {
		t.Error("Something went wrong; expected to connect!")
	}
}

func TestMqttConnect(t *testing.T) {
	mysb := defaultTestMQTT()
	mysb.onConnect(mysb.client)
}

type mockMessage struct {
	topic   string
	payload []byte
}

func (m *mockMessage) Duplicate() bool {
	return true
}

func (m *mockMessage) Qos() byte {
	return 'a'
}

func (m *mockMessage) Retained() bool {
	return true
}

func (m *mockMessage) Topic() string {
	return m.topic
}

func (m *mockMessage) MessageID() uint16 {
	return 0
}

func (m *mockMessage) Payload() []byte {
	return m.payload
}

func (m *mockMessage) Ack() {
}
