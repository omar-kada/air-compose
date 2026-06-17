package events

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogHandler is a mock implementation of slog.Handler using testify.
type MockLogHandler struct {
	mock.Mock
}

func (m *MockLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	args := m.Called(ctx, level)
	return args.Bool(0)
}

func (m *MockLogHandler) Handle(ctx context.Context, record slog.Record) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	args := m.Called(attrs)
	return args.Get(0).(slog.Handler)
}

func (m *MockLogHandler) WithGroup(name string) slog.Handler {
	args := m.Called(name)
	return args.Get(0).(slog.Handler)
}

func TestNewSlogWriter(t *testing.T) {
	writer := NewSlogWriter(slog.LevelInfo, "test")

	assert.NotNil(t, writer, "Expected non-nil writer")
	assert.Equal(t, slog.LevelInfo, writer.level, "Expected level Info")
}

func TestWrite(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a mock handler
	mockHandler := new(MockLogHandler)

	// Set up expectations
	mockHandler.On("Enabled", mock.Anything, slog.LevelInfo).Return(true)
	mockHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

	// Create a logger with the mock handler
	logger := slog.New(mockHandler)

	// Replace the default logger with our test logger
	slog.SetDefault(logger)
	defer func() {
		// Reset the default logger
		slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	}()

	// Create a SlogWriter with Info level
	writer := NewSlogWriter(slog.LevelInfo, "test")

	// Test message
	testMsg := "test message"

	// Write to the writer
	n, err := writer.Write([]byte(testMsg))

	// Check for errors
	assert.NoError(t, err, "Unexpected error")

	// Check the number of bytes written
	assert.Equal(t, len(testMsg), n, "Bytes written mismatch")

	// Check if the mock handler's methods were called as expected
	mockHandler.AssertExpectations(t)
}

func TestWriteWithDifferentLevels(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Test cases
	testCases := []struct {
		level    slog.Level
		expected slog.Level
	}{
		{slog.LevelDebug, slog.LevelDebug},
		{slog.LevelInfo, slog.LevelInfo},
		{slog.LevelWarn, slog.LevelWarn},
		{slog.LevelError, slog.LevelError},
	}

	for _, tc := range testCases {
		// Create a mock handler
		mockHandler := new(MockLogHandler)

		// Set up expectations
		mockHandler.On("Enabled", mock.Anything, tc.expected).Return(true)
		mockHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		// Create a logger with the mock handler
		logger := slog.New(mockHandler)

		// Replace the default logger with our test logger
		slog.SetDefault(logger)
		defer func() {
			// Reset the default logger
			slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
		}()
		// Create a SlogWriter with the test level
		writer := NewSlogWriter(tc.level, "test")

		// Test message
		testMsg := "test message"

		// Write to the writer
		_, _ = writer.Write([]byte(testMsg))

		// Check if the mock handler's methods were called as expected
		mockHandler.AssertExpectations(t)
	}
}
