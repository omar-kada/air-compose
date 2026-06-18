package events

import (
	"context"
	"testing"

	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/testutil"

	"github.com/stretchr/testify/assert"
)

func TestEventTransformer_HandleEvent(t *testing.T) {
	configStore := testutil.NewConfigGetter(models.Config{})
	transformer := NewEventTransformer(configStore)

	t.Run("should set IsNotification based on config", func(t *testing.T) {
		cfg := models.Config{
			Settings: models.Settings{
				Notifications: models.NotificationConfig{
					NotificationTypes: []models.EventType{models.EventMisc},
				},
			},
		}
		srcEvent := models.SourceEvent{
			Type: models.EventMisc,
		}

		configStore.Set(cfg)
		event := transformer.HandleEvent(context.Background(), srcEvent)

		assert.True(t, event.IsNotification)
	})

	t.Run("should transform source event into event with additional metadata", func(t *testing.T) {
		srcEvent := models.SourceEvent{
			Type: models.EventMisc,
			Msg:  "Test Message",
		}

		event := transformer.HandleEvent(context.Background(), srcEvent)

		assert.Equal(t, srcEvent.Type, event.Type)
		assert.Equal(t, srcEvent.Msg, event.Msg)
		assert.NotZero(t, event.Time)
	})
}
