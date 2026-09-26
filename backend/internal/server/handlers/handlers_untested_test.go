package handlers

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/storage"

	"github.com/stretchr/testify/assert"
)

// --- NotificationsAPIList ---

func TestNotificationsAPIList_Success(t *testing.T) {
	m := &Mock{}
	store, _ := config.NewConfigStore(filepath.Join(t.TempDir(), "config.yaml"), events.NewBus(1))
	assert.NoError(t, store.Update(models.Config{}))

	h := NewBusinessHandler(store, m, m, m, m, m, m, m)

	events := []models.Event{
		{ID: 1, Msg: "event1", Type: models.EventDeploymentStarted},
		{ID: 2, Msg: "event2", Type: models.EventError},
	}
	m.On("GetNotifications", storage.NewIDCursor(2, uint64(0))).Return(events, nil)

	req := api.NotificationsAPIListRequestObject{Params: api.NotificationsAPIListParams{Limit: 2}}
	resp, err := h.NotificationsAPIList(context.Background(), req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.NotificationsAPIList200JSONResponse:
		assert.Len(t, r.Items, 2)
		assert.Equal(t, uint64(1), r.Items[0].ID)
		assert.Equal(t, "event1", r.Items[0].Msg)
		assert.True(t, r.PageInfo.HasNextPage)
		assert.Equal(t, "2", r.PageInfo.EndCursor)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestNotificationsAPIList_InvalidOffset(t *testing.T) {
	m := &Mock{}
	store, _ := config.NewConfigStore(filepath.Join(t.TempDir(), "config.yaml"), events.NewBus(1))
	assert.NoError(t, store.Update(models.Config{}))

	h := NewBusinessHandler(store, m, m, m, m, m, m, m)

	offset := "notuint"
	req := api.NotificationsAPIListRequestObject{Params: api.NotificationsAPIListParams{Limit: 1, Offset: &offset}}
	resp, err := h.NotificationsAPIList(context.Background(), req)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid after value")
}

func TestNotificationsAPIList_InvalidLimit(t *testing.T) {
	m := &Mock{}
	store, _ := config.NewConfigStore(filepath.Join(t.TempDir(), "config.yaml"), events.NewBus(1))
	assert.NoError(t, store.Update(models.Config{}))

	h := NewBusinessHandler(store, m, m, m, m, m, m, m)

	req := api.NotificationsAPIListRequestObject{Params: api.NotificationsAPIListParams{Limit: 0}}
	resp, err := h.NotificationsAPIList(context.Background(), req)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid first value")
}

func TestNotificationsAPIList_GetNotificationsError(t *testing.T) {
	m := &Mock{}
	store, _ := config.NewConfigStore(filepath.Join(t.TempDir(), "config.yaml"), events.NewBus(1))
	assert.NoError(t, store.Update(models.Config{}))

	h := NewBusinessHandler(store, m, m, m, m, m, m, m)

	errStore := errors.New("db error")
	m.On("GetNotifications", storage.NewIDCursor(2, uint64(0))).Return([]models.Event{}, errStore)

	req := api.NotificationsAPIListRequestObject{Params: api.NotificationsAPIListParams{Limit: 2}}
	resp, err := h.NotificationsAPIList(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, errStore, err)

	switch r := resp.(type) {
	case api.NotificationsAPIList200JSONResponse:
		assert.Empty(t, r.Items)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestNotificationsAPIList_WithOffset(t *testing.T) {
	m := &Mock{}
	store, _ := config.NewConfigStore(filepath.Join(t.TempDir(), "config.yaml"), events.NewBus(1))
	assert.NoError(t, store.Update(models.Config{}))

	h := NewBusinessHandler(store, m, m, m, m, m, m, m)

	events := []models.Event{
		{ID: 5, Msg: "page2event", Type: models.EventDeploymentStarted, Time: time.Now()},
	}
	offset := "3"
	m.On("GetNotifications", storage.NewIDCursor(2, uint64(3))).Return(events, nil)

	req := api.NotificationsAPIListRequestObject{Params: api.NotificationsAPIListParams{Limit: 2, Offset: &offset}}
	resp, err := h.NotificationsAPIList(context.Background(), req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.NotificationsAPIList200JSONResponse:
		assert.Len(t, r.Items, 1)
		assert.Equal(t, uint64(5), r.Items[0].ID)
		assert.False(t, r.PageInfo.HasNextPage)
		assert.Equal(t, "5", r.PageInfo.EndCursor)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}
