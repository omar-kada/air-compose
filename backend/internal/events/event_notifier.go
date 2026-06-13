// Package events handles logic related to events
package events

import (
	"context"
	"fmt"
	"log/slog"

	"omar-kada/air-compose/internal/models"

	"github.com/containrrr/shoutrrr"
)

// NotificationEventHandler is an event handler that sends notifications
type NotificationEventHandler struct {
	configStore models.ConfigGetter
	eventStore  EventStorage
	Send        func(rawURL string, message string) error
}

// NewNotificationEventHandler creates a new notification event handler
func NewNotificationEventHandler(configStore models.ConfigGetter, eventStore EventStorage) Handler {
	return &NotificationEventHandler{
		configStore: configStore,
		eventStore:  eventStore,
		Send:        shoutrrr.Send,
	}
}

// HandleEvent sends a notification for the event
func (h *NotificationEventHandler) HandleEvent(_ context.Context, event models.Event) {
	cfg := h.configStore.Get()
	event.IsNotification = h.sendNotification(cfg, event)
	h.storeNotification(event)
}

func (h *NotificationEventHandler) sendNotification(cfg models.Config, event models.Event) bool {
	if !cfg.IsEventNotificationEnabled(event.Type) {
		return false
	}
	if cfg.Settings.Notifications.NotificationURL != "" {
		message := event.Type.ToEmoji() + " " + event.Type.ToText()
		if event.ObjectID != 0 {
			message += fmt.Sprintf(" - [%v] %v", event.ObjectID, event.ObjectName)
		}
		if event.Msg != "" {
			message += fmt.Sprintf(" :\n %v", event.Msg)
		}

		err := h.Send(cfg.Settings.Notifications.NotificationURL, message)
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
