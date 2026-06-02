package mocks

import (
	"omar-kada/air-compose/internal/models"

	"github.com/stretchr/testify/mock"
)

// Inspector is a mock implementation of the Inspector interface.
type Inspector struct {
	mock.Mock
}

// GetManagedStacks returns the managed stacks state.
func (m *Inspector) GetManagedStacks() (models.StacksState, error) {
	args := m.Called()
	return args.Get(0).(models.StacksState), args.Error(1)
}

// GetCurrentStacks returns the current stacks state for the given services.
func (m *Inspector) GetCurrentStacks(services []string) (models.StacksState, error) {
	args := m.Called(services)
	return args.Get(0).(models.StacksState), args.Error(1)
}
