package logs

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"omar-kada/air-compose/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultLogger(t *testing.T) {
	tests := []struct {
		name      string
		logParams models.LoggerParams
		expected  slog.Level
		humanLog  bool
	}{
		{
			name: "Human log",
			logParams: models.LoggerParams{
				LogFile:           filepath.Join(t.TempDir(), "app.log"),
				MaxLogFileSize:    10,
				MaxLogHistorySize: 100,
				LogLevel:          "info",
				HumanLog:          "true",
			},
			expected: slog.LevelInfo,
			humanLog: true,
		},
		{
			name: "JSON log",
			logParams: models.LoggerParams{
				LogFile:           filepath.Join(t.TempDir(), "app.log"),
				MaxLogFileSize:    10,
				MaxLogHistorySize: 100,
				LogLevel:          "debug",
				HumanLog:          "false",
			},
			expected: slog.LevelDebug,
			humanLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := NewHistoryHub(tt.logParams.MaxLogHistorySize)
			logger := NewDefaultLogger(hub, tt.logParams)

			assert.NotNil(t, logger)
			assert.True(t, logger.Enabled(context.Background(), tt.expected))
		})
	}
}

func TestInitLogHub(t *testing.T) {

	logFile := filepath.Join(t.TempDir(), "app.log")

	tests := []struct {
		name      string
		logParams models.LoggerParams
		expected  int
	}{
		{
			name: "With existing logs",
			logParams: models.LoggerParams{
				LogFile:           logFile,
				MaxLogFileSize:    10,
				MaxLogHistorySize: 100,
				LogLevel:          "info",
				HumanLog:          "true",
			},
			expected: 2,
		},
		{
			name: "With existing logs debug",
			logParams: models.LoggerParams{
				LogFile:           logFile,
				MaxLogFileSize:    10,
				MaxLogHistorySize: 100,
				LogLevel:          "debug",
				HumanLog:          "true",
			},
			expected: 3,
		},
		{
			name: "Without existing logs",
			logParams: models.LoggerParams{
				LogFile:           filepath.Join(t.TempDir(), "nonexistent.log"),
				MaxLogFileSize:    10,
				MaxLogHistorySize: 100,
				LogLevel:          "info",
				HumanLog:          "true",
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := strings.Join([]string{
				"{\"time\":\"2024-01-01T00:00:00Z\",\"level\":\"INFO\",\"msg\":\"one\"}",
				"{\"time\":\"2024-01-01T00:00:01Z\",\"level\":\"WARN\",\"msg\":\"two\"}",
				"",
			}, "\n")
			err := os.WriteFile(logFile, []byte(content), 0o600)
			assert.NoError(t, err)

			hub := InitLogHub(tt.logParams)
			assert.NotNil(t, hub)
			slog.Debug("three")

			client := hub.Register()
			records := waitForLogs(t, client.Send)
			assert.Len(t, records, tt.expected)
		})
	}
}

func waitForLogs(t *testing.T, ch <-chan []slog.Record) []slog.Record {
	t.Helper()

	select {
	case msg := <-ch:
		return msg
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for log snapshot")
		return nil
	}
}
