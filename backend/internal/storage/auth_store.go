package storage

import (
	"errors"
	"omar-kada/air-compose/internal/models"
)

var (
	// ErrEmptyToken indicates that the provided token is empty
	ErrEmptyToken = errors.New("empty token")
)

// AuthStore is an abstraction of all authorization related database operations
type AuthStore interface {
	UserStorage
	SessionStorage
	NewAuth(username string, token models.Token) (models.Token, error)
	RevokeToken(token models.Token) error
	GetUsernameFromToken(token models.TokenValue) string
	RevokeAllTokens(username string) error
}

// authStorage implements the Storage interface
type authStorage struct {
	UserStorage
	SessionStorage
	tokenHolder *TokenHolder
}

// NewAuthStorage creates a authorization storage
func NewAuthStorage(userStore UserStorage, sessionStore SessionStorage, tokenHolder *TokenHolder) (AuthStore, error) {
	return &authStorage{userStore, sessionStore, tokenHolder}, nil
}

func (s authStorage) NewAuth(username string, token models.Token) (models.Token, error) {
	if token.Value == "" {
		return models.Token{}, ErrEmptyToken
	}
	_, err := s.NewSession(token, username)
	if err != nil {
		return models.Token{}, err
	}
	err = s.tokenHolder.InsertToken(token.Value, username, token.Expires)
	return token, err
}

func (s authStorage) RevokeToken(token models.Token) error {
	s.tokenHolder.RemoveToken(token.Value)
	return s.RevokeRefreshToken(token.RefreshToken)
}

// GetUsernameFromToken retrieves the username associated with a token
func (s authStorage) GetUsernameFromToken(token models.TokenValue) string {
	return s.tokenHolder.GetUsernameFromToken(token)
}

func (s authStorage) RevokeAllTokens(username string) error {
	s.tokenHolder.RemoveAllUserTokens(username)
	return s.RevokeAllUserSessions(username)
}
