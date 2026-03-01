// Package events handles logic related to events
package events

import (
	"context"
	"fmt"
	"log/slog"

	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"

	"github.com/containrrr/shoutrrr"
)

// NotificationEventHandler is an event handler that sends notifications
type NotificationEventHandler struct {
	configStore storage.ConfigStore
	eventStore  storage.EventStorage
	Send        func(rawURL string, message string) error
}

// NewNotificationEventHandler creates a new notification event handler
func NewNotificationEventHandler(configStore storage.ConfigStore, eventStore storage.EventStorage) EventHandler {
	return &NotificationEventHandler{
		configStore: configStore,
		eventStore:  eventStore,
		Send:        shoutrrr.Send,
	}
}

// HandleEvent sends a notification for the event
func (h *NotificationEventHandler) HandleEvent(_ context.Context, event models.Event) {
	cfg, err := h.configStore.Get()
	if err != nil {
		slog.Error("can't retrieve config", "error", err)
		return
	}
	event.IsNotification = h.sendNotification(cfg, event)
	h.storeNotification(event)
}

func (h *NotificationEventHandler) sendNotification(cfg models.Config, event models.Event) bool {
	if !cfg.IsEventNotificationEnabled(event.Type) {
		return false
	}
	event.IsNotification = true
	if cfg.Settings.NotificationURL != "" {
		message := event.Type.ToEmoji() + " " + event.Type.ToText()
		if event.ObjectID != 0 {
			message += fmt.Sprintf(" - [%v] %v", event.ObjectID, event.ObjectName)
		}
		if event.Msg != "" {
			message += fmt.Sprintf(" :\n %v", event.Msg)
		}

		err := h.Send(cfg.Settings.NotificationURL, message)
		if err != nil {
			slog.Error("can't send notification", "error", err)
		}
	}
	return true
}

func (h *NotificationEventHandler) storeNotification(event models.Event) {
	err := h.eventStore.StoreEvent(event)
	if err != nil {
		slog.Error("can't store event", "error", err)
	}
}
