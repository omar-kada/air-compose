// Package events handles logic related to events
package events

import (
	"context"

	"omar-kada/air-compose/internal/models"
)

// EventTransformer transforms source events into events with additional metadata.
type EventTransformer struct {
	configStore models.ConfigGetter
}

// NewEventTransformer creates a new EventTransformer with the given config store.
func NewEventTransformer(configStore models.ConfigGetter) *EventTransformer {
	return &EventTransformer{
		configStore: configStore,
	}
}

// HandleEvent sends a notification for the event
func (t *EventTransformer) HandleEvent(ctx context.Context, srcEvent models.SourceEvent) models.Event {
	cfg := t.configStore.Get()
	event := models.FromSourceEvent(ctx, srcEvent)
	event.IsNotification = cfg.IsEventNotificationEnabled(event.Type)
	return event
}
