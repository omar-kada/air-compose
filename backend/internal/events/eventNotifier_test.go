package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/mock"
)

// MockSend is a mock implementation of the Send function
type MockSend struct {
	mock.Mock
}

// Send is the mock implementation of the Send function
func (m *MockSend) Send(rawURL string, message string) error {
	args := m.Called(rawURL, message)
	return args.Error(0)
}

// MockConfigStore is a mock implementation of the ConfigStore interface
type MockConfigStore struct {
	mock.Mock
	storage.ConfigStore
}

func (m *MockConfigStore) Get() (models.Config, error) {
	args := m.Called()
	return args.Get(0).(models.Config), args.Error(1)
}

func (m *MockConfigStore) IsEventNotificationEnabled(eventType models.EventType) bool {
	args := m.Called(eventType)
	return args.Bool(0)
}

// MockConfigStore is a mock implementation of the ConfigStore interface
type MockEventStore struct {
	mock.Mock
	storage.EventStorage
}

func (m *MockEventStore) StoreEvent(event models.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestNotificationEventHandler_HandleEvent(t *testing.T) {
	mockSend := new(MockSend)
	mockConfigStore := new(MockConfigStore)
	mockEventStore := new(MockEventStore)

	handler := NewNotificationEventHandler(mockConfigStore, mockEventStore).(*NotificationEventHandler)
	handler.Send = mockSend.Send

	t.Run("should send notification when config is valid and event is enabled", func(t *testing.T) {
		cfg := models.Config{
			Settings: models.Settings{
				NotificationURL:   "http://example.com",
				NotificationTypes: []models.EventType{models.EventMisc},
			},
		}
		event := models.Event{
			Type:       models.EventMisc,
			ObjectID:   1,
			ObjectName: "Test Object",
			Msg:        "Test Message",
		}

		mockConfigStore.On("Get").Return(cfg, nil)
		mockEventStore.On("StoreEvent", mock.Anything).Return(nil)

		mockSend.On("Send", cfg.Settings.NotificationURL, event.Type.ToEmoji()+" "+event.Type.ToText()+" - [1] Test Object :\n Test Message").Return(nil)

		handler.HandleEvent(context.Background(), event)

		mockConfigStore.AssertExpectations(t)
		mockSend.AssertExpectations(t)
	})

	t.Run("should not send notification when config is invalid", func(t *testing.T) {
		event := models.Event{
			Type:       models.EventMisc,
			ObjectID:   1,
			ObjectName: "Test Object",
			Msg:        "Test Message",
		}

		mockConfigStore.On("Get").Return(models.Config{}, errors.New("config error"))

		handler.HandleEvent(context.Background(), event)

		mockConfigStore.AssertExpectations(t)
		mockSend.AssertNotCalled(t, "Send")
	})

	t.Run("should not send notification when event is not enabled", func(t *testing.T) {
		cfg := models.Config{
			Settings: models.Settings{
				NotificationURL:   "http://example.com",
				NotificationTypes: []models.EventType{},
			},
		}
		event := models.Event{
			Type:       models.EventMisc,
			ObjectID:   1,
			ObjectName: "Test Object",
			Msg:        "Test Message",
		}

		mockConfigStore.On("Get").Return(cfg, nil)
		mockEventStore.On("StoreEvent", mock.Anything).Return(nil)

		handler.HandleEvent(context.Background(), event)

		mockConfigStore.AssertExpectations(t)
		mockSend.AssertNotCalled(t, "Send")
	})
}

func TestStoringEventHandler_HandleEvent(t *testing.T) {
	mockSend := new(MockSend)
	mockConfigStore := new(MockConfigStore)
	mockEventStore := new(MockEventStore)

	handler := NewNotificationEventHandler(mockConfigStore, mockEventStore).(*NotificationEventHandler)
	handler.Send = mockSend.Send

	event := models.Event{
		ID:         2,
		Type:       models.EventMisc,
		Msg:        "Test event",
		Time:       time.Now(),
		ObjectID:   1,
		ObjectName: "Test object",
	}

	mockConfigStore.On("Get").Return(models.Config{}, nil)
	mockEventStore.On("StoreEvent", event).Return(nil).Once()

	// Call the HandleEvent method with the event
	handler.HandleEvent(context.Background(), event)

	// Assert that the StoreEvent method was called with the event
	mockEventStore.AssertCalled(t, "StoreEvent", event)
}

func TestNotificationEventHandler_HandleEvent_NotificationFlag_Enabled(t *testing.T) {
	mockSend := new(MockSend)
	mockConfigStore := new(MockConfigStore)
	mockEventStore := new(MockEventStore)

	handler := NewNotificationEventHandler(mockConfigStore, mockEventStore).(*NotificationEventHandler)
	handler.Send = mockSend.Send

	cfg := models.Config{
		Settings: models.Settings{
			NotificationURL:   "http://example.com",
			NotificationTypes: []models.EventType{models.EventMisc},
		},
	}
	event := models.Event{
		Type:       models.EventMisc,
		ObjectID:   1,
		ObjectName: "Test Object",
		Msg:        "Test Message",
	}

	mockConfigStore.On("Get").Return(cfg, nil)
	mockEventStore.On("StoreEvent", mock.MatchedBy(func(e models.Event) bool {
		return e.ObjectID == 1 && e.IsNotification == true
	})).Return(nil)
	mockSend.On("Send", cfg.Settings.NotificationURL, mock.Anything).Return(nil)

	handler.HandleEvent(context.Background(), event)

	mockEventStore.AssertExpectations(t)
}

func TestNotificationEventHandler_HandleEvent_NotificationFlag_Disabled(t *testing.T) {
	mockSend := new(MockSend)
	mockConfigStore := new(MockConfigStore)
	mockEventStore := new(MockEventStore)

	handler := NewNotificationEventHandler(mockConfigStore, mockEventStore).(*NotificationEventHandler)
	handler.Send = mockSend.Send
	cfg := models.Config{
		Settings: models.Settings{
			NotificationURL:   "http://example.com",
			NotificationTypes: []models.EventType{},
		},
	}
	event := models.Event{
		Type:       models.EventMisc,
		ObjectID:   2,
		ObjectName: "Test Object",
		Msg:        "Test Message",
	}

	mockConfigStore.On("Get").Return(cfg, nil)
	mockEventStore.On("StoreEvent", mock.MatchedBy(func(e models.Event) bool {
		return e.ObjectID == 2 && e.IsNotification == false
	})).Return(nil)

	handler.HandleEvent(context.Background(), event)

	mockEventStore.AssertExpectations(t)
}
