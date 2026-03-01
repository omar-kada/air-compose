package users

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
		for i := 0; i < 100; i++ {
			th.InsertToken(token, username, expiryTime)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			th.RemoveToken(token)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			assert.Equal(t, username, th.GetUsernameFromToken(token2))
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines to finish
	<-done
	<-done
}
