package testutil

import (
	"log/slog"

	"omar-kada/air-compose/internal/storage"

	"gorm.io/gorm"
)

// NewMemoryStorage instanciates a new memory storage
func NewMemoryStorage() *gorm.DB {
	db, err := storage.NewGormDb(":memory:", 0o000)
	if err != nil {
		slog.Error("couldn't init memory store")
	}
	return db
}
