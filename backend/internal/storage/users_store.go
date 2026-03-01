package storage

import (
	"errors"
	"fmt"
	"log/slog"

	"omar-kada/air-compose/models"

	"gorm.io/gorm"
)

var (
	// ErrEmptyUsername is returned when an empty username is provided
	ErrEmptyUsername = errors.New("empty username")
	// ErrNotFound is returned when a record is not found in the database
	ErrNotFound = errors.New("not found")
)

// UserStorage is an abstraction of all user database operations
type UserStorage interface {
	HasUsers() (bool, error)
	UserByUsername(username string) (models.User, error)
	UpsertUser(user models.User) (models.User, error)
	DeleteUserByUserName(username string) (bool, error)
	NewSession(token models.Token, username string) (models.Session, error)
	SessionByRefreshToken(token models.TokenValue) (models.Session, error)
	RevokeRefreshToken(token models.TokenValue) error
	RevokeAllUserSessions(username string) error
}

// gormUserStorage implements the Storage interface using GORM
type gormUserStorage struct {
	db *gorm.DB
}

// NewUsersStorage creates a users storage and run migrations
func NewUsersStorage(db *gorm.DB) (UserStorage, error) {
	if err := db.AutoMigrate(&models.User{}, &models.Session{}); err != nil {
		return nil, err
	}
	return &gormUserStorage{db: db}, nil
}

// HasUsers checks if there are any users in the database
func (s *gormUserStorage) HasUsers() (bool, error) {
	var exists bool
	if err := s.db.Model(&models.User{}).Select("EXISTS (SELECT 1 FROM users)").Find(&exists).Error; err != nil {
		return false, err
	}
	return exists, nil
}

// UserByUsername retrieves a user by their username
func (s *gormUserStorage) UserByUsername(username string) (models.User, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, nil
		}
		return models.User{}, err
	}
	return user, nil
}

// UpsertUser updates an existing user in the database
func (s *gormUserStorage) UpsertUser(user models.User) (models.User, error) {
	if user.Username == "" {
		return models.User{}, ErrEmptyUsername
	}
	if err := s.db.Save(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *gormUserStorage) DeleteUserByUserName(username string) (bool, error) {
	tx := s.db.Where("username = ?", username).Delete(&models.User{})
	if err := tx.Error; err != nil {
		return false, err
	}
	return tx.RowsAffected > 0, nil
}

func (s *gormUserStorage) NewSession(token models.Token, username string) (models.Session, error) {
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

func (s *gormUserStorage) SessionByRefreshToken(token models.TokenValue) (models.Session, error) {
	var session models.Session
	if err := s.db.Where("refresh_token = ?", token).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Session{}, ErrNotFound
		}
		return models.Session{}, err
	}
	return session, nil
}

func (s *gormUserStorage) RevokeRefreshToken(token models.TokenValue) error {
	var session models.Session
	slog.Debug("Revoking refresh token", "refresh token", token)
	if err := s.db.Where("refresh_token = ?", token).First(&session).Error; err != nil {
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

func (s *gormUserStorage) RevokeAllUserSessions(username string) error {
	slog.Debug(fmt.Sprintf("Revoking all serssions for user '%s'", username))

	if err := s.db.Model(&models.Session{}).Where("username = ?", username).Update("revoked", true).Error; err != nil {
		return err
	}
	return nil
}
