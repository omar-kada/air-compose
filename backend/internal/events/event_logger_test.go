package events

import (
	"context"
	"log/slog"
	"reflect"
	"testing"

	"omar-kada/air-compose/internal/models"
)

func TestLoggingEventHandler_HandleEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    models.Event
		expected string
	}{
		{
			name: "basic event",
			event: models.Event{
				Type: "test",
				Msg:  "testMessage",
			},
			expected: "[EVENT] test: testMessage",
		},
		{
			name: "object event",
			event: models.Event{
				Type:       "test",
				ObjectID:   123,
				ObjectName: "testObject",
				Msg:        "testMessage",
			},
			expected: "[EVENT] #123 - test: testMessage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewLoggingEventHandler()
			mockLogger := &mockLogger{}
			slog.SetDefault(slog.New(mockLogger))

			handler.HandleEvent(context.Background(), tt.event)

			if len(mockLogger.loggedEvents) != 1 {
				t.Errorf("expected 1 logged event, got %d", len(mockLogger.loggedEvents))
			}

			if !reflect.DeepEqual(mockLogger.loggedEvents[0].Message, tt.expected) {
				t.Errorf("logged event doesn't match expected:\nExpected: %+v\nActual: %+v",
					tt.expected, mockLogger.loggedEvents[0])
			}
		})
	}
}

// mockLogger is a mock implementation of slog.Handler for testing
type mockLogger struct {
	loggedEvents []slog.Record
}

func (*mockLogger) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (m *mockLogger) Handle(_ context.Context, record slog.Record) error {
	attrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	m.loggedEvents = append(m.loggedEvents, slog.Record{
		Level:   record.Level,
		Message: record.Message,
	})
	return nil
}

func (m *mockLogger) WithAttrs(_ []slog.Attr) slog.Handler {
	return m
}

func (m *mockLogger) WithGroup(_ string) slog.Handler {
	return m
}
