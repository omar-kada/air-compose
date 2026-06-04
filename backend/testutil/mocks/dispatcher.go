package mocks

import (
	"context"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"

	"github.com/stretchr/testify/mock"
)

// Dispatcher is a mock implementation of the Dispatcher interface.
type Dispatcher struct {
	mock.Mock
}

// Dispatch records the call to the Dispatch method with the provided arguments.
func (m *Dispatcher) Dispatch(ctx context.Context, eventType models.EventType, message string) {
	m.Called(ctx, eventType, message)
}

// AddHandler registers an event handler with the dispatcher.
func (m *Dispatcher) AddHandler(handler events.Handler) {
	m.Called(handler)
}
