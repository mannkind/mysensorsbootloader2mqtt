package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/mannkind/twomqtt"
	log "github.com/sirupsen/logrus"
)

var moqClient = twomqtt.MoqClient{}

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

	c := initialize()
	var tests = []struct {
		Request       string
		Response      string
		AutoIDEnabled bool
	}{
		{fmt.Sprintf("%s/255/255/3/0/3", c.config.SubTopic), "13", true},
		{fmt.Sprintf("%s/255/255/3/0/3", c.config.SubTopic), "", false},
	}
	for _, v := range tests {
		msg := &twomqtt.MoqMessage{
			TopicSrc:   v.Request,
			PayloadSrc: "",
		}

		c := initialize()
		c.config.NextID = 12
		c.config.AutoIDEnabled = v.AutoIDEnabled
		matched := c.idRequest(moqClient, msg)

		actual := matched.Payload
		expected := v.Response
		if actual != expected {
			t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", actual, expected)
		}
	}
}

func TestMqttConfigurationRequest(t *testing.T) {
	setEnvs()
	defer clearEnvs()

	c := initialize()
	c.config.FirmwareBasePath = "test_files"

	msg := &twomqtt.MoqMessage{
		TopicSrc:   fmt.Sprintf("%s/1/255/4/0/0", c.config.SubTopic),
		PayloadSrc: "010001005000D446",
	}

	matched := c.configurationRequest(moqClient, msg)

	actual := matched.Payload
	expected := "010001005000D446"
	if actual != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", actual, expected)
	}
}

func TestMqttDataRequest(t *testing.T) {
	setEnvs()
	defer clearEnvs()

	c := initialize()
	c.config.FirmwareBasePath = "test_files"

	msg := &twomqtt.MoqMessage{
		TopicSrc:   fmt.Sprintf("%s/1/255/4/0/2", c.config.SubTopic),
		PayloadSrc: "010001000100",
	}

	matched := c.dataRequest(moqClient, msg)

	actual := matched.Payload
	expected := "0100010001000C946E000C946E000C946E000C946E00"
	if actual != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", actual, expected)
	}
}

func TestMqttBootloaderCommand(t *testing.T) {
	setEnvs()
	defer clearEnvs()

	c := initialize()
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

		c.bootloaderCommand(moqClient, msg)
		if _, ok := c.config.BootloaderCommands[v.To]; !ok {
			t.Error("Bootloader command not found")
		} else {
			if ok, matched := c.runBootloaderCommand(moqClient, v.To); !ok {
				t.Error("Bootloader command not run")
			} else {
				actual := matched.Payload
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

	c := initialize()
	if ok, _ := c.runBootloaderCommand(moqClient, "1"); ok {
		t.Error("Bootloader command didn't exist, should not have returned true")
	}
}
