// Package users provides user management and authentication services.
package users

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"
)

var (
	// ErrAlreadyRegistered indicates that a user is already registered
	ErrAlreadyRegistered = errors.New("already registered")
	// ErrUserNotFound indicates that a user was not found
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidPassword indicates that the provided password is incorrect
	ErrInvalidPassword = errors.New("invalid password")
	// ErrInvalidRefreshToken indicates that the provided refreshtoken is incorrect
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// use in memory storage for token , and user a timer to remove it from the map once expired,
//

// Service abstracts authorization operations
type Service interface {
	AuthService
	AccountService
}

// AuthService abstracts authentication operations
type AuthService interface {
	Login(credentials models.Credentials) (models.Token, error)
	Register(credentials models.Credentials) (models.Token, error)
	IsRegistered() (bool, error)
	Logout(token models.Token) error
	GetUsernameByToken(token models.Token) (string, error)
	RefreshToken(token models.Token) (models.Token, error)
}

// AccountService abstracts account management operations
type AccountService interface {
	GetUser(username string) (models.User, error)
	DeleteUser(username string) (bool, error)
	ChangePassword(username string, oldPass string, newPass string) (bool, error)
}

type service struct {
	userStore   storage.UserStorage
	tokenHolder TokenHolder
}

// NewService creates a new userService
func NewService(userStore storage.UserStorage) Service {
	return &service{
		userStore:   userStore,
		tokenHolder: *NewTokenHolder(),
	}
}

// Login authenticates a user and returns their authentication token.
func (a *service) Login(credentials models.Credentials) (models.Token, error) {
	user, err := a.userStore.UserByUsername(credentials.Username)
	if err != nil {
		return models.Token{}, fmt.Errorf("error finding user: %w", err)
	}

	if user.Username == "" {
		return models.Token{}, ErrUserNotFound
	}

	if !checkPasswordHash(credentials.Password, user.HashedPassword) {
		return models.Token{}, ErrInvalidPassword
	}

	return a.insertNewToken(user.Username)
}

// IsRegistered checks if any users are registered in the system.
func (a *service) IsRegistered() (bool, error) {
	hasUsers, err := a.userStore.HasUsers()
	if err != nil {
		return false, fmt.Errorf("error checking if users exist: %w", err)
	}
	return hasUsers, nil
}

// Register creates a new user account with the provided credentials
func (a *service) Register(credentials models.Credentials) (models.Token, error) {
	hasUsers, err := a.userStore.HasUsers()
	if err != nil {
		return models.Token{}, err
	}
	if hasUsers {
		return models.Token{}, ErrAlreadyRegistered
	}
	hashedPassword, err := hashPassword(credentials.Password)
	if err != nil {
		return models.Token{}, fmt.Errorf("error hashing password: %w", err)
	}

	user := models.User{
		Username:       credentials.Username,
		HashedPassword: hashedPassword,
	}
	_, err = a.userStore.UpsertUser(user)
	if err != nil {
		return models.Token{}, fmt.Errorf("error creating user: %w", err)
	}

	return a.insertNewToken(user.Username)
}

func (a *service) insertNewToken(username string) (models.Token, error) {
	token, err := generateToken()
	if err != nil {
		return models.Token{}, err
	}

	_, err = a.userStore.NewSession(token, username)
	if err != nil {
		return models.Token{}, err
	}
	a.tokenHolder.InsertToken(token.Value, username, token.Expires)
	return token, err
}

func (a *service) RefreshToken(token models.Token) (models.Token, error) {
	session, err := a.userStore.SessionByRefreshToken(token.RefreshToken)
	if err != nil {
		return models.Token{}, err
	}
	valid := !session.Revoked && session.RefreshExpires.After(time.Now())
	if !valid {
		slog.Warn("invalid session", "session", session)
		err = a.userStore.RevokeAllUserSessions(session.Username)
		if err != nil {
			return models.Token{}, err
		}
		return models.Token{}, ErrInvalidRefreshToken
	}

	err = a.userStore.RevokeRefreshToken(token.RefreshToken)
	if err != nil {
		return models.Token{}, err
	}

	a.tokenHolder.RemoveToken(token.Value)
	return a.insertNewToken(session.Username)
}

// Logout invalidates the user's authentication token.
func (a *service) Logout(token models.Token) error {
	if a.tokenHolder.GetUsernameFromToken(token.Value) == "" {
		return ErrUserNotFound
	}

	err := a.userStore.RevokeRefreshToken(token.RefreshToken)
	if err != nil {
		return err
	}
	a.tokenHolder.RemoveToken(token.Value)

	return nil
}

// GetUsernameByToken retrieves a username by their authentication token.
func (a *service) GetUsernameByToken(token models.Token) (string, error) {
	username := a.tokenHolder.GetUsernameFromToken(token.Value)
	if username == "" {
		return "", ErrUserNotFound
	}
	return username, nil
}

// GetUser retrieves a user by their username.
func (a *service) GetUser(username string) (models.User, error) {
	user, err := a.userStore.UserByUsername(username)
	if err == nil && user.Username == "" {
		return user, ErrUserNotFound
	}
	return user, err
}

// DeleteUser removes a user from the system by their username.
func (a *service) DeleteUser(username string) (bool, error) {
	return a.userStore.DeleteUserByUserName(username)
}

// ChangePassword updates a user's password after verifying the old password.
func (a *service) ChangePassword(username string, oldPass string, newPass string) (bool, error) {
	user, err := a.userStore.UserByUsername(username)
	if err != nil {
		return false, fmt.Errorf("error finding user: %w", err)
	}

	if user.Username == "" {
		return false, ErrUserNotFound
	}

	if !checkPasswordHash(oldPass, user.HashedPassword) {
		return false, ErrInvalidPassword
	}

	hashedPassword, err := hashPassword(newPass)
	if err != nil {
		return false, fmt.Errorf("error hashing new password: %w", err)
	}

	user.HashedPassword = hashedPassword
	if _, err := a.userStore.UpsertUser(user); err != nil {
		return false, fmt.Errorf("error updating user: %w", err)
	}

	return true, nil
}
