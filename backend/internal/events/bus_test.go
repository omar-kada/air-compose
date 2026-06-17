package events

import (
	"context"
	"testing"
	"time"

	"omar-kada/air-compose/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventHandler is a mock implementation of the Handler interface for testing.
type MockEventHandler struct {
	mock.Mock
}

// HandleEvent implements the Handler interface.
func (m *MockEventHandler) HandleEvent(ctx context.Context, event models.Event) {
	m.Called(ctx, event)
}

// TestNewBus tests the creation of a new Bus instance.
func TestNewBus(t *testing.T) {
	bufferSize := 10

	bus := NewBus(bufferSize)

	assert.NotNil(t, bus, "Expected non-nil bus")
	assert.Equal(t, bufferSize, cap(bus.ch), "Expected buffer size to match")
}

// TestRegister tests the registration of handlers.
func TestRegister(t *testing.T) {

	bus := NewBus(10)
	handler1 := new(MockEventHandler)
	handler2 := new(MockEventHandler)

	bus.Register(handler1, handler2)

	assert.Len(t, bus.handlers, 2, "Expected two handlers to be registered")
}

// TestRun tests the running of the bus.
func TestRun(t *testing.T) {
	bus := NewBus(10)
	handler := new(MockEventHandler)
	done := make(chan struct{})
	handler.On("HandleEvent", mock.Anything, mock.Anything).Return().Run(func(_ mock.Arguments) {
		close(done)
	})

	bus.Register(handler)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go bus.Run(ctx)
	srcEvent := models.SourceEvent{
		Type: models.EventMisc,
		Msg:  "test message",
	}

	bus.Publish(ctx, srcEvent)

	select {
	case <-done:
		// handler called
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for handler to be called")
	}

	handler.AssertCalled(t, "HandleEvent", ctx, mock.MatchedBy(func(event models.Event) bool {
		return event.Type == srcEvent.Type && event.Msg == srcEvent.Msg
	}))
}

// TestDispatchOverflow tests that events are dropped when the buffer is full.
func TestDispatchOverflow(t *testing.T) {

	bus := NewBus(1) // Small buffer to force overflow
	handler := new(MockEventHandler)
	bus.Register(handler)

	// Fill the buffer
	srcEvent := models.SourceEvent{
		Type: models.EventMisc,
		Msg:  "test message",
	}
	bus.Publish(context.Background(), srcEvent)

	// This should overflow
	handler.On("HandleEvent", mock.Anything, mock.Anything).Return()
	bus.Publish(context.Background(), srcEvent)

	// Start the bus to process events
	ctx := t.Context()
	go bus.Run(ctx)

	// Verify only the first event was processed
	time.Sleep(100 * time.Millisecond)

	handler.AssertNumberOfCalls(t, "HandleEvent", 1)
	handler.AssertCalled(t, "HandleEvent", mock.Anything, mock.MatchedBy(func(event models.Event) bool {
		return event.Type == srcEvent.Type && event.Msg == srcEvent.Msg
	}))
}

// TestDispatchOverflow tests that events are dropped when the buffer is full.
func TestPublishWaith(t *testing.T) {
	bus := NewBus(1) // Small buffer to force overflow
	handler := new(MockEventHandler)
	bus.Register(handler)

	// Fill the buffer
	srcEvent := models.SourceEvent{
		Type: models.EventMisc,
		Msg:  "test message",
	}
	bus.Publish(context.Background(), srcEvent)

	// This should overflow
	handler.On("HandleEvent", mock.Anything, mock.Anything).Return()
	go bus.PublishWait(context.Background(), srcEvent)

	// Start the bus to process events
	ctx := t.Context()
	go bus.Run(ctx)

	// Verify only the first event was processed
	time.Sleep(100 * time.Millisecond)

	handler.AssertNumberOfCalls(t, "HandleEvent", 2)
}

func TestPanicHandler(t *testing.T) {
	bus := NewBus(10)
	panicHandler := HandlerFunc(func(_ context.Context, _ models.Event) {
		panic("test panic")
	})
	bus.Register(panicHandler)

	normalHandler := new(MockEventHandler)
	normalHandler.On("HandleEvent", mock.Anything, mock.Anything).Return()
	bus.Register(normalHandler)

	ctx := t.Context()

	go bus.Run(ctx)

	// Publish an event that will cause the panic handler to panic
	srcEvent := models.SourceEvent{
		Type: models.EventMisc,
		Msg:  "test message",
	}
	bus.Publish(ctx, srcEvent)

	// Wait for the handler to panic and recover
	time.Sleep(100 * time.Millisecond)

	normalHandler.AssertCalled(t, "HandleEvent", mock.Anything, mock.MatchedBy(func(event models.Event) bool {
		return event.Type == srcEvent.Type && event.Msg == srcEvent.Msg
	}))
}
