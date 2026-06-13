package events

import (
	"context"
	"fmt"
	"log/slog"

	"omar-kada/air-compose/internal/models"
)

// NewLoggingEventHandler creates a new logging event handler
func NewLoggingEventHandler() Handler {
	return HandlerFunc(func(ctx context.Context, event models.Event) {
		if event.ObjectID != 0 {
			slog.Log(ctx, slog.LevelInfo, fmt.Sprintf("[EVENT] #%v - %v: %v", event.ObjectID, event.Type, event.Msg))
		} else {
			slog.Log(ctx, slog.LevelInfo, fmt.Sprintf("[EVENT] %v: %v", event.Type, event.Msg))
		}
	})
}
