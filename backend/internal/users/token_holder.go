package users

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"omar-kada/air-compose/internal/models"
)

var (
	// ErrTokenExpired indicates that the token being processed has expired
	ErrTokenExpired = errors.New("token is expired")
)

// TokenHolder manages a map of tokens with automatic expiration handling
type TokenHolder struct {
	tokens map[models.TokenValue]string
	rwMu   sync.RWMutex
}

// NewTokenHolder creates a new TokenHolder instance
func NewTokenHolder() *TokenHolder {
	return &TokenHolder{
		tokens: make(map[models.TokenValue]string),
	}
}

// InsertToken adds a token to the holder with an expiration time
func (th *TokenHolder) InsertToken(token models.TokenValue, username string, expiryTime time.Time) error {
	th.rwMu.Lock()
	defer th.rwMu.Unlock()

	if time.Now().After(expiryTime) {
		slog.Warn("inserting expired token")
		return ErrTokenExpired
	}

	th.tokens[token] = username
	time.AfterFunc(time.Until(expiryTime), func() {
		th.RemoveToken(token)
	})
	return nil
}

// RemoveToken removes a token from the holder
func (th *TokenHolder) RemoveToken(token models.TokenValue) {
	th.rwMu.Lock()
	defer th.rwMu.Unlock()

	delete(th.tokens, token)
}

// RemoveAllUserTokens removes all tokens associated with the given username
func (th *TokenHolder) RemoveAllUserTokens(username string) {
	th.rwMu.Lock()
	defer th.rwMu.Unlock()

	// Create a slice of tokens to delete
	var tokensToDelete []models.TokenValue
	for token, user := range th.tokens {
		if user == username {
			tokensToDelete = append(tokensToDelete, token)
		}
	}

	// Delete the tokens outside the loop
	for _, token := range tokensToDelete {
		delete(th.tokens, token)
	}
}

// GetUsernameFromToken retrieves the username associated with a token
func (th *TokenHolder) GetUsernameFromToken(token models.TokenValue) string {
	th.rwMu.RLock()
	defer th.rwMu.RUnlock()

	return th.tokens[token]
}
