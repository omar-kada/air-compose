package testutil

import (
	"testing"

	"omar-kada/air-compose/internal/storage"

	"gorm.io/gorm"
)

// NewMemoryStorage instanciates a new memory storage
func NewMemoryStorage(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := storage.NewGormDb(":memory:", 0o000)
	if err != nil {
		t.Fatalf("couldn't init memory store %v", err)
	}
	return db
}
