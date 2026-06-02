package events

import (
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/storage"

	"gorm.io/gorm"
)

// gormEventStorage implements the Storage interface using GORM
type gormEventStorage struct {
	db *gorm.DB
}

// EventStorage is an abstraction of all event database operations
type EventStorage interface {
	StoreEvent(event models.Event) error
	GetEvents(objectID uint64) ([]models.Event, error)
	GetNotifications(c storage.Cursor[uint64]) ([]models.Event, error)
}

// NewEventStorage creates a storage for events using gorm
func NewEventStorage(db *gorm.DB) (EventStorage, error) {
	// Auto-migrate models types
	if err := db.AutoMigrate(&models.Event{}); err != nil {
		return nil, err
	}
	return &gormEventStorage{db: db}, nil
}

// StoreEvent creates a new event and associates it with an existing deployment
func (s *gormEventStorage) StoreEvent(event models.Event) error {
	if err := s.db.Create(&event).Error; err != nil {
		return err
	}
	return nil
}

// GetEvents retrieves all events associated with a specific object ID
func (s *gormEventStorage) GetEvents(objectID uint64) ([]models.Event, error) {
	var event []models.Event
	if err := s.db.Where("object_id = ?", objectID).Find(&event).Error; err != nil {
		return nil, err
	}
	return event, nil
}

// GetNotifications retrieves all events that are notifications
func (s *gormEventStorage) GetNotifications(c storage.Cursor[uint64]) ([]models.Event, error) {
	var notifs []models.Event
	if err := s.db.
		Scopes(storage.Paginate(c)).Order("Time desc").Where("is_notification = true").Find(&notifs).Error; err != nil {
		return nil, err
	}
	return notifs, nil
}
