package main

import (
	"fmt"
	"testing"

	mqttExtCfg "github.com/mannkind/paho.mqtt.golang.ext/cfg"
	mqttExtDI "github.com/mannkind/paho.mqtt.golang.ext/di"
	"gopkg.in/yaml.v2"
)

const nodeRequestHex = "test_files/1/1/firmware.hex"

func defaultTestMQTT() *mqttClient {
	var testConfig = `
      nodes:
        default: { type: 1, version: 1 }
        1: { type: 1, version: 1, queueMessages: true }
    `

	mysensorsbootloader2mqtt := newMqttClient(NewConfig(mqttExtCfg.NewMQTTConfig()), mqttExtDI.NewMQTTFuncWrapper())
	mysensorsbootloader2mqtt.autoIDEnabled = true
	mysensorsbootloader2mqtt.nextID = 12
	mysensorsbootloader2mqtt.firmwareBasePath = "test_files"
	if err := yaml.Unmarshal([]byte(testConfig), &mysensorsbootloader2mqtt); err != nil {
		panic(err)
	}
	return mysensorsbootloader2mqtt
}

func TestMqttIDRequest(t *testing.T) {
	mysensorsbootloader2mqtt := defaultTestMQTT()
	var tests = []struct {
		Request       string
		Response      string
		AutoIDEnabled bool
	}{
		{fmt.Sprintf("%s/255/255/3/0/3", mysensorsbootloader2mqtt.subTopic), fmt.Sprintf("%s/255/255/3/0/4 %s", mysensorsbootloader2mqtt.pubTopic, "13"), true},
		{fmt.Sprintf("%s/255/255/3/0/3", mysensorsbootloader2mqtt.subTopic), "", false},
	}
	for _, v := range tests {
		msg := &mockMessage{
			topic:   v.Request,
			payload: []byte(""),
		}

		expected := v.Response

		mysensorsbootloader2mqtt.lastPublished = ""
		mysensorsbootloader2mqtt.autoIDEnabled = v.AutoIDEnabled
		mysensorsbootloader2mqtt.idRequest(mysensorsbootloader2mqtt.client, msg)
		if mysensorsbootloader2mqtt.lastPublished != expected {
			t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", mysensorsbootloader2mqtt.lastPublished, expected)
		}
	}
}

func TestMqttConfigurationRequest(t *testing.T) {
	mysensorsbootloader2mqtt := defaultTestMQTT()
	msg := &mockMessage{
		topic:   fmt.Sprintf("%s/1/255/4/0/0", mysensorsbootloader2mqtt.subTopic),
		payload: []byte("010001005000D446"),
	}

	expected := fmt.Sprintf("%s/1/255/4/0/1 %s", mysensorsbootloader2mqtt.pubTopic, "010001005000D446")
	mysensorsbootloader2mqtt.configurationRequest(mysensorsbootloader2mqtt.client, msg)
	if mysensorsbootloader2mqtt.lastPublished != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", mysensorsbootloader2mqtt.lastPublished, expected)
	}
}

func TestMqttDataRequest(t *testing.T) {
	mysensorsbootloader2mqtt := defaultTestMQTT()
	msg := &mockMessage{
		topic:   fmt.Sprintf("%s/1/255/4/0/0", mysensorsbootloader2mqtt.subTopic),
		payload: []byte("010001000100"),
	}

	expected := fmt.Sprintf("%s/1/255/4/0/3 %s", mysensorsbootloader2mqtt.pubTopic, "0100010001000C946E000C946E000C946E000C946E00")

	mysensorsbootloader2mqtt.dataRequest(mysensorsbootloader2mqtt.client, msg)
	if mysensorsbootloader2mqtt.lastPublished != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", mysensorsbootloader2mqtt.lastPublished, expected)
	}
}

func TestMqttBootloaderCommand(t *testing.T) {
	mysensorsbootloader2mqtt := defaultTestMQTT()
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

		mysensorsbootloader2mqtt.bootloaderCommand(mysensorsbootloader2mqtt.client, msg)
		if _, ok := mysensorsbootloader2mqtt.bootloaderCommands[v.To]; !ok {
			t.Error("Bootloader command not found")
		} else {
			if ok := mysensorsbootloader2mqtt.runBootloaderCommand(mysensorsbootloader2mqtt.client, v.To); !ok {
				t.Error("Bootloader command not run")
			} else {
				expected := fmt.Sprintf("%s/%s/255/4/0/1 %s", mysensorsbootloader2mqtt.pubTopic, v.To, v.ExpectedPayload)
				if mysensorsbootloader2mqtt.lastPublished != expected {
					t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", mysensorsbootloader2mqtt.lastPublished, expected)
				}
			}
		}
	}
}

func TestMqttBadBootloaderCommand(t *testing.T) {
	mysensorsbootloader2mqtt := defaultTestMQTT()
	if ok := mysensorsbootloader2mqtt.runBootloaderCommand(mysensorsbootloader2mqtt.client, "1"); ok {
		t.Error("Bootloader command didn't exist, should not have returned true")
	}
}

func TestMqttConnect(t *testing.T) {
	mysensorsbootloader2mqtt := defaultTestMQTT()
	mysensorsbootloader2mqtt.onConnect(mysensorsbootloader2mqtt.client)
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
