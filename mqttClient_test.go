package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/mannkind/twomqtt"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.PanicLevel)
}

func setEnvs() {
	nodes := `
nodes:
  default: { type: 1, version: 1 }
  1: { type: 1, version: 1, queueMessages: true }
`

	os.Setenv("MYSENSORS_NODES", nodes)
}

func clearEnvs() {
	os.Setenv("MYSENSORS_NODES", "")
}

func TestMqttIDRequest(t *testing.T) {
	setEnvs()
	defer clearEnvs()

	mysensorsbootloader2mqtt := initialize()
	var tests = []struct {
		Request       string
		Response      string
		AutoIDEnabled bool
	}{
		{fmt.Sprintf("%s/255/255/3/0/3", mysensorsbootloader2mqtt.SubTopic), "13", true},
		{fmt.Sprintf("%s/255/255/3/0/3", mysensorsbootloader2mqtt.SubTopic), "", false},
	}
	for _, v := range tests {
		msg := &twomqtt.MoqMessage{
			TopicSrc:   v.Request,
			PayloadSrc: "",
		}

		mysensorsbootloader2mqtt := initialize()
		mysensorsbootloader2mqtt.NextID = 12
		mysensorsbootloader2mqtt.AutoIDEnabled = v.AutoIDEnabled
		mysensorsbootloader2mqtt.idRequest(mysensorsbootloader2mqtt.Client, msg)

		actual := mysensorsbootloader2mqtt.LastPublishedOnTopic(fmt.Sprintf("%s/255/255/3/0/4", mysensorsbootloader2mqtt.PubTopic))
		expected := v.Response
		if actual != expected {
			t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", actual, expected)
		}
	}
}

func TestMqttConfigurationRequest(t *testing.T) {
	setEnvs()
	defer clearEnvs()

	mysensorsbootloader2mqtt := initialize()
	mysensorsbootloader2mqtt.FirmwareBasePath = "test_files"

	msg := &twomqtt.MoqMessage{
		TopicSrc:   fmt.Sprintf("%s/1/255/4/0/0", mysensorsbootloader2mqtt.SubTopic),
		PayloadSrc: "010001005000D446",
	}

	mysensorsbootloader2mqtt.configurationRequest(mysensorsbootloader2mqtt.Client, msg)

	actual := mysensorsbootloader2mqtt.LastPublishedOnTopic(fmt.Sprintf("%s/1/255/4/0/1", mysensorsbootloader2mqtt.PubTopic))
	expected := "010001005000D446"
	if actual != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", actual, expected)
	}
}

func TestMqttDataRequest(t *testing.T) {
	setEnvs()
	defer clearEnvs()

	mysensorsbootloader2mqtt := initialize()
	mysensorsbootloader2mqtt.FirmwareBasePath = "test_files"

	msg := &twomqtt.MoqMessage{
		TopicSrc:   fmt.Sprintf("%s/1/255/4/0/2", mysensorsbootloader2mqtt.SubTopic),
		PayloadSrc: "010001000100",
	}

	mysensorsbootloader2mqtt.dataRequest(mysensorsbootloader2mqtt.Client, msg)

	actual := mysensorsbootloader2mqtt.LastPublishedOnTopic(fmt.Sprintf("%s/1/255/4/0/3", mysensorsbootloader2mqtt.PubTopic))
	expected := "0100010001000C946E000C946E000C946E000C946E00"
	if actual != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", actual, expected)
	}
}

func TestMqttBootloaderCommand(t *testing.T) {
	setEnvs()
	defer clearEnvs()

	mysensorsbootloader2mqtt := initialize()
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
		msg := &twomqtt.MoqMessage{
			TopicSrc:   fmt.Sprintf("mysensors/bootloader/%s/%s", v.To, v.Cmd),
			PayloadSrc: v.Payload,
		}

		mysensorsbootloader2mqtt.bootloaderCommand(mysensorsbootloader2mqtt.Client, msg)
		if _, ok := mysensorsbootloader2mqtt.BootloaderCommands[v.To]; !ok {
			t.Error("Bootloader command not found")
		} else {
			if ok := mysensorsbootloader2mqtt.runBootloaderCommand(mysensorsbootloader2mqtt.Client, v.To); !ok {
				t.Error("Bootloader command not run")
			} else {
				actual := mysensorsbootloader2mqtt.LastPublishedOnTopic(fmt.Sprintf("%s/%s/255/4/0/1", mysensorsbootloader2mqtt.PubTopic, v.To))
				expected := v.ExpectedPayload
				if actual != expected {
					t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", actual, expected)
				}
			}
		}
	}
}

func TestMqttBadBootloaderCommand(t *testing.T) {
	setEnvs()
	defer clearEnvs()

	mysensorsbootloader2mqtt := initialize()
	if ok := mysensorsbootloader2mqtt.runBootloaderCommand(mysensorsbootloader2mqtt.Client, "1"); ok {
		t.Error("Bootloader command didn't exist, should not have returned true")
	}
}

func TestMqttConnect(t *testing.T) {
	setEnvs()
	defer clearEnvs()

	mysensorsbootloader2mqtt := initialize()
	mysensorsbootloader2mqtt.onConnect(mysensorsbootloader2mqtt.Client)
}
