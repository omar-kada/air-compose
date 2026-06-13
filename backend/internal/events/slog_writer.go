package events

import (
	"context"
	"log/slog"
)

// SlogWriter is a writer that writes to slog.
type SlogWriter struct {
	logger *slog.Logger
	level  slog.Level
	id     string
}

// NewSlogWriter creates a new SlogWriter.
func NewSlogWriter(level slog.Level, id string) *SlogWriter {
	return &SlogWriter{logger: slog.Default(), level: level, id: id}
}

// Write implements the io.Writer interface.
func (sw *SlogWriter) Write(p []byte) (n int, err error) {
	sw.logger.Log(context.Background(), sw.level, sw.id, slog.String("msg", string(p)))
	return len(p), nil
}
