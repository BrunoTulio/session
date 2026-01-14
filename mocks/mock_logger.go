package mocks

import (
	"github.com/BrunoTulio/session"
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)

}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)

}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func NewMockLogger() session.Logger {
	return &MockLogger{}
}
