package users

import (
	"testing"
	"time"

	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/storage"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupAuthStorage(t *testing.T) (AuthStore, *gorm.DB) {
	db, err := storage.NewGormDb(":memory:", 0o000)
	if err != nil {
		t.Fatalf("couldn't init memory store %v", err)
	}
	userStore, err := NewUsersStorage(db)
	assert.NoError(t, err)
	sessionStore, err := NewSessionStorage(db)
	assert.NoError(t, err)
	tokenHolder := NewTokenHolder()
	authStore, err := NewAuthStorage(userStore, sessionStore, tokenHolder)
	assert.NoError(t, err)
	return authStore, db
}

func TestNewAuth(t *testing.T) {
	s, _ := setupAuthStorage(t)

	// Create a test token
	token := models.Token{
		Value:   "test_token",
		Expires: time.Now().Add(time.Hour),
	}

	// Create new auth
	_, err := s.NewAuth("testuser", token)
	assert.NoError(t, err)

	// Verify the token was created
	username := s.GetUsernameFromToken(token.Value)
	assert.Equal(t, "testuser", username)
}

func TestRevokeToken(t *testing.T) {
	s, _ := setupAuthStorage(t)

	// Create a test token
	token := models.Token{
		Value:   "test_token",
		Expires: time.Now().Add(time.Hour),
	}

	// Create new auth
	_, err := s.NewAuth("testuser", token)
	assert.NoError(t, err)

	// Revoke the token
	err = s.RevokeToken(token)
	assert.NoError(t, err)

	// Verify the token was revoked
	username := s.GetUsernameFromToken(token.Value)
	assert.Empty(t, username)
}

func TestGetUsernameFromToken(t *testing.T) {
	s, _ := setupAuthStorage(t)

	// Test non-existent token
	username := s.GetUsernameFromToken("nonexistent_token")
	assert.Empty(t, username)

	// Create a test token
	token := models.Token{
		Value:   "test_token",
		Expires: time.Now().Add(time.Hour),
	}

	// Create new auth
	_, err := s.NewAuth("testuser", token)
	assert.NoError(t, err)

	// Test existing token
	username = s.GetUsernameFromToken(token.Value)
	assert.Equal(t, "testuser", username)
}

func TestRevokeAllSessions(t *testing.T) {
	s, _ := setupAuthStorage(t)

	// Create test tokens for the same user
	token1 := models.Token{
		Value:   "token1",
		Expires: time.Now().Add(time.Hour),
	}
	token2 := models.Token{
		Value:   "token2",
		Expires: time.Now().Add(time.Hour),
	}
	_, err := s.NewAuth("testuser", token1)
	assert.NoError(t, err)
	_, err = s.NewAuth("testuser", token2)
	assert.NoError(t, err)

	// Revoke all sessions for the user
	err = s.RevokeAllTokens("testuser")
	assert.NoError(t, err)

	// Verify both tokens were revoked
	username := s.GetUsernameFromToken("token1")
	assert.Empty(t, username)

	username = s.GetUsernameFromToken("token2")
	assert.Empty(t, username)
}
func TestNewAuth_EmptyToken(t *testing.T) {
	s, _ := setupAuthStorage(t)

	// Test with empty token
	emptyToken := models.Token{}
	_, err := s.NewAuth("testuser", emptyToken)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyToken, err)
}

func TestRevokeToken_NonExistentToken(t *testing.T) {
	s, _ := setupAuthStorage(t)

	// Test revoking non-existent token
	nonExistentToken := models.Token{
		Value:   "nonexistent_token",
		Expires: time.Now().Add(time.Hour),
	}
	err := s.RevokeToken(nonExistentToken)
	assert.ErrorIs(t, err, ErrNotFound)

	// Verify no error occurred
	username := s.GetUsernameFromToken(nonExistentToken.Value)
	assert.Empty(t, username)
}

func TestGetUsernameFromToken_ExpiredToken(t *testing.T) {
	s, _ := setupAuthStorage(t)

	// Create an expired token
	expiredToken := models.Token{
		Value:   "expired_token",
		Expires: time.Now().Add(-time.Hour),
	}
	_, err := s.NewAuth("testuser", expiredToken)
	assert.ErrorIs(t, err, ErrTokenExpired)

	// Verify the token is still retrievable (tokenHolder handles expiration)
	username := s.GetUsernameFromToken(expiredToken.Value)
	assert.Empty(t, username)
}
