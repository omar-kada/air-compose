package mocks

import (
	"context"
	"omar-kada/air-compose/internal/models"

	"github.com/stretchr/testify/mock"
)

// EventPublisher is a mock implementation of the EventPublisher interface.
type EventPublisher struct {
	mock.Mock
}

// Publish records the call to the Publich method with the provided arguments.
func (m *EventPublisher) Publish(ctx context.Context, srcEvent models.SourceEvent) {
	m.Called(ctx, srcEvent)
}
