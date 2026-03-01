package users

import (
	"sync"
	"time"

	"omar-kada/air-compose/models"
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
func (th *TokenHolder) InsertToken(token models.TokenValue, username string, expiryTime time.Time) {
	th.rwMu.Lock()
	defer th.rwMu.Unlock()

	th.tokens[token] = username
	time.AfterFunc(time.Until(expiryTime), func() {
		th.RemoveToken(token)
	})
}

// RemoveToken removes a token from the holder
func (th *TokenHolder) RemoveToken(token models.TokenValue) {
	th.rwMu.Lock()
	defer th.rwMu.Unlock()

	delete(th.tokens, token)
}

// GetUsernameFromToken retrieves the username associated with a token
func (th *TokenHolder) GetUsernameFromToken(token models.TokenValue) string {
	th.rwMu.RLock()
	defer th.rwMu.RUnlock()

	return th.tokens[token]
}
