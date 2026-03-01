package storage

import (
	"testing"
	"time"

	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupEventStorage(t *testing.T) (EventStorage, *gorm.DB) {
	db, err := NewGormDb(":memory:", 0o000)
	if err != nil {
		t.Fatalf("new db: %v", err)
	}
	eventStore, err := NewEventStorage(db)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	return eventStore, db
}

func TestEventStorage_Migrates(t *testing.T) {
	_, db := setupEventStorage(t)
	// ensure migrations created the deployments table
	has := db.Migrator().HasTable(&models.Event{})
	assert.True(t, has)
}

func initDeployment(t *testing.T, db *gorm.DB, title string) models.Deployment {
	dep := models.Deployment{
		Title:  title,
		Author: "author",
		Diff:   "",
		Status: models.DeploymentStatusRunning,
		Time:   time.Now(),
	}
	if err := db.Create(&dep).Error; err != nil {
		t.Fatalf("couldn't init deployment : %v", err)
	}
	return dep
}

func TestStoreEventAndGetEvents(t *testing.T) {
	s, db := setupEventStorage(t)
	dep := initDeployment(t, db, "title1")

	ev := models.Event{Type: models.EventMisc, Msg: "ok", ObjectID: dep.ID}
	assert.NoError(t, s.StoreEvent(ev))
	events, err := s.GetEvents(dep.ID)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "ok", events[0].Msg)
}

func TestStoreEvent_NoDeployment(t *testing.T) {
	s, _ := setupEventStorage(t)

	ev := models.Event{Type: models.EventMisc, Msg: "ok", ObjectID: 1}
	assert.Error(t, s.StoreEvent(ev))
	events, err := s.GetEvents(1)
	assert.NoError(t, err)
	assert.Len(t, events, 0)
}
