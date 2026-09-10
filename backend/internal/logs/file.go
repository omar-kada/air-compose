package logs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"omar-kada/air-compose/internal/models"
	"os"
	"strings"
	"time"
)

// LoadRecordsFromFile reads JSON log entries from a slog JSON file and reconstructs
// the original slog.Record values so they can be replayed into the in-memory hub.
func LoadRecordsFromFile(path string) ([]slog.Record, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read log file: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(bs))
	records := make([]slog.Record, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			return nil, fmt.Errorf("decode log entry: %w", err)
		}

		record, err := recordFromJSON(payload)
		if err != nil {
			return nil, err
		}
		if record.Message != "" {
			records = append(records, record)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan log file: %w", err)
	}

	return records, nil
}

func recordFromJSON(payload map[string]any) (slog.Record, error) {
	timestamp, ok := payload["time"].(string)
	if !ok {
		timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	parsedTime, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		parsedTime = time.Now()
	}

	levelValue, ok := payload["level"].(string)
	if !ok {
		levelValue = slog.LevelInfo.String()
	}
	level := models.ParseLogLevel(levelValue)

	message, _ := payload["msg"].(string)
	record := slog.NewRecord(parsedTime, level, message, 0)

	attrs := make([]slog.Attr, 0, len(payload))
	for key, value := range payload {
		switch key {
		case "time", "level", "msg":
			continue
		default:
			attrs = append(attrs, slog.Any(key, value))
		}
	}
	if len(attrs) > 0 {
		record.AddAttrs(attrs...)
	}

	return record, nil
}
