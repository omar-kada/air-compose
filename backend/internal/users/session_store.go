package users

import (
	"errors"
	"fmt"
	"log/slog"

	"omar-kada/air-compose/internal/models"

	"gorm.io/gorm"
)

var (
	// ErrNotFound is returned when a record is not found in the database
	ErrNotFound = errors.New("not found")
)

// SessionStorage is an abstraction of all session database operations
type SessionStorage interface {
	NewSession(token models.Token, username string) (models.Session, error)
	SessionByRefreshToken(token models.TokenValue) (models.Session, error)
	RevokeRefreshToken(token models.TokenValue) error
	RevokeAllUserSessions(username string) error
}

// gormSessionStorage implements the Storage interface using GORM
type gormSessionStorage struct {
	db *gorm.DB
}

// NewSessionStorage creates a session storage and run migrations
func NewSessionStorage(db *gorm.DB) (SessionStorage, error) {
	if err := db.AutoMigrate(&models.Session{}); err != nil {
		return nil, err
	}
	return &gormSessionStorage{db: db}, nil
}

func (s *gormSessionStorage) NewSession(token models.Token, username string) (models.Session, error) {
	session := models.Session{
		RefreshToken:   string(token.RefreshToken),
		RefreshExpires: token.RefreshExpires,
		Revoked:        false,
		Username:       username,
	}

	if err := s.db.Save(&session).Error; err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (s *gormSessionStorage) SessionByRefreshToken(token models.TokenValue) (models.Session, error) {
	var session models.Session
	if err := s.db.Where("refresh_token = ?", token).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Session{}, ErrNotFound
		}
		return models.Session{}, err
	}
	return session, nil
}

func (s *gormSessionStorage) RevokeRefreshToken(tokenValue models.TokenValue) error {
	var session models.Session
	slog.Debug("Revoking refresh token", "refresh token", tokenValue)
	if err := s.db.Where("refresh_token = ?", tokenValue).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	session.Revoked = true
	if err := s.db.Save(&session).Error; err != nil {
		return err
	}
	return nil
}

func (s *gormSessionStorage) RevokeAllUserSessions(username string) error {
	slog.Debug(fmt.Sprintf("Revoking all serssions for user '%s'", username))

	if err := s.db.Model(&models.Session{}).Where("username = ?", username).Update("revoked", true).Error; err != nil {
		return err
	}
	return nil
}
