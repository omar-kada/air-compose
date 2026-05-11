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
	// ErrInvalidLoginMethod indicates that a user it trying to login with an invalid method
	ErrInvalidLoginMethod = errors.New("invalid login method")
	// ErrEmptyToken indicates that the provided token is empty
	ErrEmptyToken = errors.New("empty token")
	// ErrInvalidRefreshToken indicates that the provided refreshtoken is incorrect
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	// ErrUserDataCorrupted indicates that user or session information are not consistent
	ErrUserDataCorrupted = errors.New("user data corrupted")
)

// Service abstracts authorization operations
type Service interface {
	AuthService
	AccountService
}

// AuthService abstracts authentication operations
type AuthService interface {
	Login(credentials models.Credentials) (models.Token, error)
	Register(credentials models.Credentials) (models.Token, error)
	Logout(token models.Token) error
	GetUsernameByToken(token models.Token) (string, error)
	RefreshToken(token models.Token) (models.Token, error)
}

// AccountService abstracts account management operations
type AccountService interface {
	IsRegistered() (bool, error)
	GetUser(username string) (models.User, error)
	DeleteUser(username string) (bool, error)
	ChangePassword(username string, oldPass string, newPass string) (bool, error)
}

type service struct {
	authStore storage.AuthStore
}

// NewService creates a new userService
func NewService(authStore storage.AuthStore) Service {
	return &service{
		authStore: authStore,
	}
}

// Login authenticates a user and returns their authentication token.
func (a *service) Login(credentials models.Credentials) (models.Token, error) {
	user, err := a.authStore.UserByUsername(credentials.Username)
	if err != nil {
		return models.Token{}, fmt.Errorf("error finding user: %w", err)
	}

	if user.Username == "" {
		return models.Token{}, ErrUserNotFound
	}

	if !checkPasswordHash(credentials.Password, user.HashedPassword) {
		return models.Token{}, ErrInvalidPassword
	}

	if user.Type != models.UserTypeLocal {
		return models.Token{}, ErrInvalidLoginMethod
	}

	return a.authStore.NewAuth(user.Username, generateToken())

}

// IsRegistered checks if any users are registered in the system.
func (a *service) IsRegistered() (bool, error) {
	hasUsers, err := a.authStore.HasUsers()
	if err != nil {
		return false, fmt.Errorf("error checking if users exist: %w", err)
	}
	return hasUsers, nil
}

// Register creates a new user account with the provided credentials
func (a *service) Register(credentials models.Credentials) (models.Token, error) {
	hasUsers, err := a.authStore.HasUsers()
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
		Type:           models.UserTypeLocal,
	}
	_, err = a.authStore.UpsertUser(user)
	if err != nil {
		return models.Token{}, fmt.Errorf("error creating user: %w", err)
	}

	return a.authStore.NewAuth(user.Username, generateToken())
}

func (a *service) RefreshToken(token models.Token) (models.Token, error) {
	session, err := a.authStore.SessionByRefreshToken(token.RefreshToken)
	if err != nil {
		return models.Token{}, err
	}
	valid := !session.Revoked && session.RefreshExpires.After(time.Now())
	if !valid {
		slog.Warn("invalid session", "session", session)
		err = a.authStore.RevokeAllTokens(session.Username)
		if err != nil {
			return models.Token{}, err
		}
		return models.Token{}, ErrInvalidRefreshToken
	}

	err = a.authStore.RevokeToken(token)
	if err != nil {
		return models.Token{}, err
	}
	return a.authStore.NewAuth(session.Username, generateToken())
}

// Logout invalidates the user's authentication token.
func (a *service) Logout(token models.Token) error {
	if a.authStore.GetUsernameFromToken(token.Value) == "" {
		return ErrUserNotFound
	}
	return a.authStore.RevokeToken(token)
}

// GetUsernameByToken retrieves a username by their authentication token.
func (a *service) GetUsernameByToken(token models.Token) (string, error) {
	username := a.authStore.GetUsernameFromToken(token.Value)
	if username == "" {
		return "", ErrUserNotFound
	}
	return username, nil
}

// GetUser retrieves a user by their username.
func (a *service) GetUser(username string) (models.User, error) {
	user, err := a.authStore.UserByUsername(username)
	if err == nil && user.Username == "" {
		return user, ErrUserNotFound
	}
	return user, err
}

// DeleteUser removes a user from the system by their username.
func (a *service) DeleteUser(username string) (bool, error) {
	return a.authStore.DeleteUserByUserName(username)
}

// ChangePassword updates a user's password after verifying the old password.
func (a *service) ChangePassword(username string, oldPass string, newPass string) (bool, error) {
	user, err := a.authStore.UserByUsername(username)
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
	if _, err := a.authStore.UpsertUser(user); err != nil {
		return false, fmt.Errorf("error updating user: %w", err)
	}

	return true, nil
}
