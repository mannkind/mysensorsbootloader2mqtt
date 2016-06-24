package transport

import (
	"fmt"
	"gopkg.in/mqtt.v0"
	"testing"
)

const nodeRequestHex = "../test_files/1/1/firmware.hex"

var testClient = new(mqtt.Client)

func defaultTestMQTT() *MQTT {
	return NewMQTT("_test_config.yaml")
}

func TestMqttIDRequest(t *testing.T) {
	myMQTT := defaultTestMQTT()
	msg := &mockMessage{
		topic:   fmt.Sprintf("%s/255/255/3/0/3", myMQTT.Settings.SubTopic),
		payload: []byte(""),
	}

	expected := fmt.Sprintf("%s/255/255/3/0/4 %s", myMQTT.Settings.PubTopic, "13")

	myMQTT.idRequest(testClient, msg)
	if myMQTT.LastPublished != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", myMQTT.LastPublished, expected)
	}
}

func TestMqttConfigurationRequest(t *testing.T) {
	myMQTT := defaultTestMQTT()
	msg := &mockMessage{
		topic:   fmt.Sprintf("%s/1/255/4/0/0", myMQTT.Settings.SubTopic),
		payload: []byte("010001005000D446"),
	}

	expected := fmt.Sprintf("%s/1/255/4/0/1 %s", myMQTT.Settings.PubTopic, "010001005000D446")
	myMQTT.configurationRequest(testClient, msg)
	if myMQTT.LastPublished != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", myMQTT.LastPublished, expected)
	}
}

func TestMqttDataRequest(t *testing.T) {
	myMQTT := defaultTestMQTT()
	msg := &mockMessage{
		topic:   fmt.Sprintf("%s/1/255/4/0/0", myMQTT.Settings.SubTopic),
		payload: []byte("010001000100"),
	}

	expected := fmt.Sprintf("%s/1/255/4/0/3 %s", myMQTT.Settings.PubTopic, "0100010001000C946E000C946E000C946E000C946E00")

	myMQTT.dataRequest(testClient, msg)
	if myMQTT.LastPublished != expected {
		t.Errorf("Wrong topic or payload - Actual: %s, Expected: %s", myMQTT.LastPublished, expected)
	}
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
