package logs

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadRecordsFromFile2(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
		err      bool
	}{
		{
			name:     "Valid log file",
			content:  "{\"time\":\"2024-01-01T00:00:00Z\",\"level\":\"INFO\",\"msg\":\"one\"}\n{\"time\":\"2024-01-01T00:00:01Z\",\"level\":\"WARN\",\"msg\":\"two\"}\n",
			expected: 2,
			err:      false,
		},
		{
			name:     "Empty file",
			content:  "",
			expected: 0,
			err:      false,
		},
		{
			name:     "Malformed JSON",
			content:  "{\"time\":\"2024-01-01T00:00:00Z\",\"level\":\"INFO\",\"msg\":\"one\"}\n{\"invalid\"=\"json\"}\n",
			expected: 0,
			err:      true,
		},
		{
			name:     "invalid values JSON",
			content:  "{\"time\":\"2024-01-01T00:00:00Z\",\"level\":\"INFO\",\"msg\":\"one\"}\n{\"invalid\":\"json\"}\n",
			expected: 1,
			err:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "app.log")
			err := os.WriteFile(path, []byte(tt.content), 0o600)
			assert.NoError(t, err)

			records, err := LoadRecordsFromFile(path)

			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, records, tt.expected)
			}
		})
	}
}

func TestLoadRecordsFromNonExistentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.log")
	records, err := LoadRecordsFromFile(path)

	assert.NoError(t, err)
	assert.Nil(t, records)
}

func TestRecordFromJSON(t *testing.T) {
	tests := []struct {
		name     string
		payload  map[string]any
		expected slog.Level
		message  string
	}{
		{
			name: "Valid payload",
			payload: map[string]any{
				"time":  "2024-01-01T00:00:00Z",
				"level": "INFO",
				"msg":   "test message",
			},
			expected: slog.LevelInfo,
			message:  "test message",
		},
		{
			name: "Missing time",
			payload: map[string]any{
				"level": "WARN",
				"msg":   "test message",
			},
			expected: slog.LevelWarn,
			message:  "test message",
		},
		{
			name: "Invalid time format",
			payload: map[string]any{
				"time":  "invalid-time",
				"level": "ERROR",
				"msg":   "test message",
			},
			expected: slog.LevelError,
			message:  "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := recordFromJSON(tt.payload)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, record.Level)
			assert.Equal(t, tt.message, record.Message)
		})
	}
}
