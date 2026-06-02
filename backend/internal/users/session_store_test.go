package users

import (
	"testing"
	"time"

	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/storage"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupSessionStorage(t *testing.T) (SessionStorage, *gorm.DB) {
	db, err := storage.NewGormDb(":memory:", 0o000)
	if err != nil {
		t.Fatalf("new db: %v", err)
	}
	sessionStore, err := NewSessionStorage(db)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	return sessionStore, db
}

func TestSessionStorage_Migrates(t *testing.T) {
	_, db := setupSessionStorage(t)
	// ensure migrations created the deployments table
	has := db.Migrator().HasTable(&models.Session{})
	assert.True(t, has)
}

func TestNewSession(t *testing.T) {
	s, _ := setupSessionStorage(t)

	// Create a test token
	token := models.Token{
		RefreshToken:   "test_refresh_token",
		RefreshExpires: time.Now().Add(time.Hour),
	}

	// Create a new session
	session, err := s.NewSession(token, "testuser")
	assert.NoError(t, err)
	assert.Equal(t, "test_refresh_token", session.RefreshToken)
	assert.Equal(t, token.RefreshExpires, session.RefreshExpires)
	assert.Equal(t, "testuser", session.Username)
	assert.False(t, session.Revoked)

	// Verify the session was created
	storedSession, err := s.SessionByRefreshToken("test_refresh_token")
	assert.NoError(t, err)
	assert.Equal(t, session.SessionID, storedSession.SessionID)
}

func TestSessionByRefreshToken(t *testing.T) {
	s, _ := setupSessionStorage(t)

	// Test non-existent token
	_, err := s.SessionByRefreshToken("nonexistent_token")
	assert.ErrorIs(t, err, ErrNotFound)

	// Create a test session
	token := models.Token{
		RefreshToken:   "test_refresh_token",
		RefreshExpires: time.Now().Add(time.Hour),
	}
	_, err = s.NewSession(token, "testuser")
	assert.NoError(t, err)

	// Test existing token
	session, err := s.SessionByRefreshToken("test_refresh_token")
	assert.NoError(t, err)
	assert.Equal(t, "test_refresh_token", session.RefreshToken)
	assert.Equal(t, "testuser", session.Username)
}

func TestRevokeRefreshToken(t *testing.T) {
	s, _ := setupSessionStorage(t)

	// Test non-existent token
	err := s.RevokeRefreshToken("nonexistent_token")
	assert.ErrorIs(t, err, ErrNotFound)

	// Create a test session
	token := models.Token{
		RefreshToken:   "test_refresh_token",
		RefreshExpires: time.Now().Add(time.Hour),
	}
	_, err = s.NewSession(token, "testuser")
	assert.NoError(t, err)

	// Revoke the token
	err = s.RevokeRefreshToken("test_refresh_token")
	assert.NoError(t, err)

	// Verify the token was revoked
	session, err := s.SessionByRefreshToken("test_refresh_token")
	assert.NoError(t, err)
	assert.True(t, session.Revoked)
}

func TestRevokeAllUserSessions(t *testing.T) {
	s, _ := setupSessionStorage(t)

	// Create test sessions for the same user
	token1 := models.Token{
		RefreshToken:   "token1",
		RefreshExpires: time.Now().Add(time.Hour),
	}
	token2 := models.Token{
		RefreshToken:   "token2",
		RefreshExpires: time.Now().Add(time.Hour),
	}
	_, err := s.NewSession(token1, "testuser")
	assert.NoError(t, err)
	_, err = s.NewSession(token2, "testuser")
	assert.NoError(t, err)

	// Revoke all sessions for the user
	err = s.RevokeAllUserSessions("testuser")
	assert.NoError(t, err)

	// Verify both sessions were revoked
	session1, err := s.SessionByRefreshToken("token1")
	assert.NoError(t, err)
	assert.True(t, session1.Revoked)

	session2, err := s.SessionByRefreshToken("token2")
	assert.NoError(t, err)
	assert.True(t, session2.Revoked)
}
