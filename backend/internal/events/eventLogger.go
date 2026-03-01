package events

import (
	"context"
	"fmt"
	"log/slog"

	"omar-kada/air-compose/models"
)

// LoggingEventHandler is an event handler that logs events
type LoggingEventHandler struct{}

// NewLoggingEventHandler creates a new logging event handler
func NewLoggingEventHandler() EventHandler {
	return &LoggingEventHandler{}
}

// HandleEvent logs the event
func (*LoggingEventHandler) HandleEvent(ctx context.Context, event models.Event) {
	slog.Log(ctx, slog.LevelInfo, fmt.Sprintf("[%v - %v] %v : %v", event.Type, event.ObjectID, event.ObjectName, event.Msg))
}
