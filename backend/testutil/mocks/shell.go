// Package mocks provides mock implementations for testing purposes.
package mocks

import (
	"omar-kada/air-compose/internal/shell"

	"github.com/stretchr/testify/mock"
)

// Executor is a mock implementation of a shell executor.
type Executor struct {
	mock.Mock
}

// Exec executes a command with given arguments and returns the output or an error.
func (m *Executor) Exec(cmd string, args ...string) ([]byte, error) {
	res := m.Called(cmd, args)
	return res.Get(0).([]byte), res.Error(1)
}

// NoLogs returns a new executor that suppresses logging.
func (m *Executor) NoLogs() shell.Executor {
	res := m.Called()
	return res.Get(0).(shell.Executor)
}
