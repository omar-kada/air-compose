// Package storage provides functionality for database operations and storage management.
package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"omar-kada/air-compose/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewGormDb creates a new instance of gorm database and runs migrations
func NewGormDb(dbFile string, addPerm os.FileMode) (*gorm.DB, error) {
	if dbFile != ":memory:" {
		if _, err := os.Stat(dbFile); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(dbFile), 0o700|addPerm); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
	}
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Warn, // Log level
			Colorful:      true,        // Enable color
		},
	)

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}
	// pragmas and pooling
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA foreign_keys = ON;")
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("couldn't init sqlite db %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// Auto-migrate models types
	if err := db.AutoMigrate(&models.Deployment{}, &models.FileDiff{}, &models.Event{}, &models.User{}); err != nil {
		return nil, err
	}
	return db, nil
}
