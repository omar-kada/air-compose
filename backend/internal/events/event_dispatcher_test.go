package events

import (
	"context"
	"testing"
	"time"

	"omar-kada/air-compose/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewDefaultDispatcher(t *testing.T) {
	eventHandlers := []Handler{NewLoggingEventHandler()}
	dispatcher := NewDefaultDispatcher(eventHandlers)

	if dispatcher == nil {
		t.Error("Expected non-nil dispatcher")
	}
}

type MockEventHandler struct {
	mock.Mock
}

func (m *MockEventHandler) HandleEvent(ctx context.Context, event models.Event) {
	m.Called(ctx, event)
}

func TestHandleEvent(t *testing.T) {
	mockHandler := new(MockEventHandler)
	done := make(chan struct{})
	mockHandler.On("HandleEvent", mock.Anything, mock.Anything).
		Return().
		Run(func(_ mock.Arguments) { close(done) })
	dispatcher := NewDefaultDispatcher([]Handler{mockHandler})

	ctx := context.Background()
	ctx = context.WithValue(ctx, objectIDCtxKey, uint64(1))
	ctx = context.WithValue(ctx, objectNameCtxKey, "test")

	eventType := models.EventMisc
	msg := "test message"

	dispatcher.Dispatch(ctx, eventType, msg)

	select {
	case <-done:
		// handler called
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for handler to be called")
	}

	mockHandler.AssertCalled(t, "HandleEvent", ctx, mock.MatchedBy(func(event models.Event) bool {
		return event.Type == eventType && event.Msg == msg
	}))
}

func TestGetDeploymentContext(t *testing.T) {
	deployment := models.Deployment{
		ID:    1,
		Title: "test deployment",
	}

	ctx := context.Background()
	newCtx := GetDeploymentContext(ctx, deployment)
	objectID, objectName := GetObjectFromContext(newCtx)

	assert.Equal(t, uint64(1), objectID)
	assert.Equal(t, "test deployment", objectName)
}

func TestNewVoidDispatcher(t *testing.T) {
	dispatcher := NewVoidDispatcher()

	if dispatcher == nil {
		t.Error("Expected non-nil dispatcher")
	}
}

func TestAddHandler(t *testing.T) {
	mockHandler := new(MockEventHandler)
	done := make(chan struct{})
	mockHandler.On("HandleEvent", mock.Anything, mock.Anything).
		Return().
		Run(func(_ mock.Arguments) { close(done) })
	dispatcher := NewDefaultDispatcher([]Handler{})
	dispatcher.AddHandler(mockHandler)
	ctx := context.Background()
	ctx = context.WithValue(ctx, objectIDCtxKey, uint64(1))
	ctx = context.WithValue(ctx, objectNameCtxKey, "test")

	eventType := models.EventMisc
	msg := "test message"

	dispatcher.Dispatch(ctx, eventType, msg)

	select {
	case <-done:
		// handler called
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for handler to be called")
	}

	mockHandler.AssertCalled(t, "HandleEvent", ctx, mock.MatchedBy(func(event models.Event) bool {
		return event.Type == eventType && event.Msg == msg
	}))
}
