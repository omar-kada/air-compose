// Package mocks provides mock implementations for testing purposes.
package mocks

import (
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
