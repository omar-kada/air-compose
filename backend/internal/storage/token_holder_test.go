package storage

import (
	"testing"
	"time"

	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
)

func TestTokenHolder(t *testing.T) {
	th := NewTokenHolder()

	// Test InsertToken and GetUsernameFromToken
	token := models.TokenValue("test-token")
	username := "test-user"
	expiryTime := time.Now().Add(1 * time.Minute)

	th.InsertToken(token, username, expiryTime)
	gotUsername := th.GetUsernameFromToken(token)

	if gotUsername != username {
		t.Errorf("GetUsernameFromToken() = %v, want %v", gotUsername, username)
	}

	// Test RemoveToken
	th.RemoveToken(token)
	gotUsername = th.GetUsernameFromToken(token)

	if gotUsername != "" {
		t.Errorf("GetUsernameFromToken() after removal = %v, want empty string", gotUsername)
	}

	// Test automatic expiration
	token = models.TokenValue("expired-token")
	th.InsertToken(token, username, time.Now().Add(-1*time.Minute))
	time.Sleep(10 * time.Millisecond) // Wait for expiration
	gotUsername = th.GetUsernameFromToken(token)

	if gotUsername != "" {
		t.Errorf("Token should have expired, but still exists")
	}
}

func TestTokenHolderConcurrency(t *testing.T) {
	th := NewTokenHolder()
	token := models.TokenValue("concurrent-token")
	token2 := models.TokenValue("concurrent-token2")
	username := "concurrent-user"
	expiryTime := time.Now().Add(1 * time.Minute)
	th.InsertToken(token2, username, expiryTime)

	// Concurrently insert and remove tokens
	done := make(chan bool)
	go func() {
		for range 100 {
			th.InsertToken(token, username, expiryTime)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for range 100 {
			th.RemoveToken(token)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for range 100 {
			assert.Equal(t, username, th.GetUsernameFromToken(token2))
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines to finish
	<-done
	<-done
}
func TestTokenHolderInsertExpiredToken(t *testing.T) {
	th := NewTokenHolder()

	token := models.TokenValue("expired-token")
	username := "test-user"
	err := th.InsertToken(token, username, time.Now().Add(-1*time.Minute))
	assert.ErrorIs(t, err, ErrTokenExpired)
	assert.Empty(t, th.GetUsernameFromToken(token))
}

func TestTokenHolderInsertEmptyUsername(t *testing.T) {
	th := NewTokenHolder()
	token := models.TokenValue("empty-username-token")
	err := th.InsertToken(token, "", time.Now().Add(1*time.Minute))
	assert.NoError(t, err)
	assert.Empty(t, th.GetUsernameFromToken(token))
}

func TestTokenHolderInsertEmptyTokenValue(t *testing.T) {
	th := NewTokenHolder()

	username := "test-user"
	err := th.InsertToken("", username, time.Now().Add(1*time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, username, th.GetUsernameFromToken(""))
}

func TestTokenHolderRemoveNonExistentToken(t *testing.T) {
	th := NewTokenHolder()
	th.RemoveToken(models.TokenValue("non-existent-token"))
	assert.Empty(t, th.GetUsernameFromToken(models.TokenValue("non-existent-token")))
}

func TestTokenHolderRemoveAllUserTokens(t *testing.T) {
	th := NewTokenHolder()
	token1 := models.TokenValue("user-token1")
	token2 := models.TokenValue("user-token2")
	username := "multi-token-user"
	th.InsertToken(token1, username, time.Now().Add(1*time.Minute))
	th.InsertToken(token2, username, time.Now().Add(1*time.Minute))
	th.RemoveAllUserTokens(username)
	assert.Empty(t, th.GetUsernameFromToken(token1))
	assert.Empty(t, th.GetUsernameFromToken(token2))
}

func TestTokenHolderConcurrentExpiration(t *testing.T) {
	th := NewTokenHolder()
	token := models.TokenValue("concurrent-expiry-token")
	username := "test-user"
	th.InsertToken(token, username, time.Now().Add(10*time.Millisecond))
	time.Sleep(20 * time.Millisecond)
	assert.Empty(t, th.GetUsernameFromToken(token))
}
