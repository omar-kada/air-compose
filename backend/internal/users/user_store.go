package users

import (
	"errors"

	"omar-kada/air-compose/internal/models"

	"gorm.io/gorm"
)

var (
	// ErrEmptyUsername is returned when an empty username is provided
	ErrEmptyUsername = errors.New("empty username")
)

// UserStorage is an abstraction of all user database operations
type UserStorage interface {
	HasUsers() (bool, error)
	UserByUsername(username string) (models.User, error)
	UpsertUser(user models.User) (models.User, error)
	DeleteUserByUserName(username string) (bool, error)
}

// gormUserStorage implements the Storage interface using GORM
type gormUserStorage struct {
	db *gorm.DB
}

// NewUsersStorage creates a users storage and run migrations
func NewUsersStorage(db *gorm.DB) (UserStorage, error) {
	if err := db.AutoMigrate(&models.User{}); err != nil {
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
