package mappers

import (
	"testing"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
)

func TestEventMapper_Map(t *testing.T) {
	// Setup
	eventMapper := EventMapper{}

	// Test data
	event := models.Event{
		ID:   1,
		Time: time.Now(),
		Msg:  "testEvent",
		Type: models.EventError,
	}

	// Expected result
	expected := api.Event{
		ID:   event.ID,
		Time: event.Time,
		Msg:  event.Msg,
		Type: api.EventType(event.Type),
	}

	// Execute
	actual := eventMapper.Map(event)

	// Assert
	assert.Equal(t, expected, actual)
}

func TestEventMapper_MapToPageInfo(t *testing.T) {
	// Setup
	eventMapper := EventMapper{}

	// Test data
	events := []models.Event{
		{
			ID:   1,
			Time: time.Now(),
			Msg:  "testEvent1",
			Type: models.EventError,
		},
		{
			ID:   2,
			Time: time.Now(),
			Msg:  "testEvent2",
			Type: models.EventError,
		},
	}

	// Test cases
	tests := []struct {
		name     string
		events   []models.Event
		limit    int
		expected api.PageInfo
	}{
		{
			name:   "No events",
			events: []models.Event{},
			limit:  2,
			expected: api.PageInfo{
				HasNextPage: false,
				EndCursor:   "",
			},
		},
		{
			name:   "Less events than limit",
			events: events,
			limit:  3,
			expected: api.PageInfo{
				HasNextPage: false,
				EndCursor:   "2",
			},
		},
		{
			name:   "Equal events to limit",
			events: events,
			limit:  2,
			expected: api.PageInfo{
				HasNextPage: true,
				EndCursor:   "2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			actual := eventMapper.MapToPageInfo(tt.events, tt.limit)

			// Assert
			assert.Equal(t, tt.expected, actual)
		})
	}
}
