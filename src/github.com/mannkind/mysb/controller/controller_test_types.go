package controller

import (
	"github.com/stretchr/testify/mock"
)

type mockMessage struct {
	mock.Mock
}

func (m mockMessage) Duplicate() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m mockMessage) Qos() byte {
	m.Called()
	return byte('a')
}

func (m mockMessage) Retained() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m mockMessage) Topic() string {
	args := m.Called()
	if s, ok := args.Get(0).(string); ok {
		return s
	}
	return ""
}

func (m mockMessage) MessageID() uint16 {
	args := m.Called()
	if s, ok := args.Get(0).(uint16); ok {
		return s
	}
	return uint16(0)
}

func (m mockMessage) Payload() []byte {
	args := m.Called()
	if s, ok := args.Get(0).([]byte); ok {
		return s
	}
	return []byte("")
}
