package users

import (
	"net/url"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOidcService_GetAuthURL(t *testing.T) {
	// Setup
	mockServer := testutil.NewOidcTestServerWithToken(t)
	defer mockServer.Close()

	userStore, _ := storage.NewUsersStorage(testutil.NewMemoryStorage(t))
	sessionStore, _ := storage.NewSessionStorage(testutil.NewMemoryStorage(t))
	tokenHolder := storage.NewTokenHolder()
	authStore, _ := storage.NewAuthStorage(userStore, sessionStore, tokenHolder)

	oidcConfig := models.OidcConfig{
		IssuerURL: mockServer.IssuerURL,
		ClientID:  testutil.ClientID,
	}

	oidcService := NewOidcService(oidcConfig, authStore)

	// Test
	redirectURL := "http://localhost:8080/callback"
	state := "teststate"
	nonce := "testnonce"

	authURL, err := oidcService.GetAuthURL(redirectURL, state, nonce)

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, authURL, mockServer.IssuerURL)
	assert.Contains(t, authURL, "response_type=code")
	assert.Contains(t, authURL, "client_id="+testutil.ClientID)
	assert.Contains(t, authURL, "redirect_uri="+url.QueryEscape(redirectURL))
	assert.Contains(t, authURL, "state="+state)
	assert.Contains(t, authURL, "nonce="+nonce)
}

func TestOidcService_LoginOidc(t *testing.T) {
	// Setup
	mockServer := testutil.NewOidcTestServerWithToken(t)
	defer mockServer.Close()

	userStore, _ := storage.NewUsersStorage(testutil.NewMemoryStorage(t))
	sessionStore, _ := storage.NewSessionStorage(testutil.NewMemoryStorage(t))
	tokenHolder := storage.NewTokenHolder()
	authStore, _ := storage.NewAuthStorage(userStore, sessionStore, tokenHolder)

	oidcConfig := models.OidcConfig{
		IssuerURL: mockServer.IssuerURL,
		ClientID:  testutil.ClientID,
	}

	oidcService := NewOidcService(oidcConfig, authStore)

	// Test
	code := mockServer.SignIDToken(testutil.ClientID, testutil.User, map[string]any{
		"email":        testutil.Email,
		"redirect-uri": "callback-url",
	})

	token, err := oidcService.LoginOidc(code, testutil.Nonce, "callback-url")

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token.Value)
	assert.NotEmpty(t, token.RefreshToken)
	// Verify the user was created
	user, err := userStore.UserByUsername(testutil.Email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, testutil.Email, user.Username)

}

func TestOidcService_LoginOidc_NonceMismatch(t *testing.T) {
	// Setup
	mockServer := testutil.NewOidcTestServerWithToken(t)
	defer mockServer.Close()

	userStore, _ := storage.NewUsersStorage(testutil.NewMemoryStorage(t))
	sessionStore, _ := storage.NewSessionStorage(testutil.NewMemoryStorage(t))
	tokenHolder := storage.NewTokenHolder()
	authStore, _ := storage.NewAuthStorage(userStore, sessionStore, tokenHolder)

	oidcConfig := models.OidcConfig{
		IssuerURL: mockServer.IssuerURL,
		ClientID:  testutil.ClientID,
	}

	oidcService := NewOidcService(oidcConfig, authStore)

	// Test
	code := mockServer.SignIDToken(testutil.ClientID, testutil.User, map[string]any{
		"email":        testutil.Email,
		"redirect-uri": "callback-url",
	})

	// Use a different nonce than what was expected
	token, err := oidcService.LoginOidc(code, "wrong-nonce", "")

	// Assert

	assert.Error(t, err)
	assert.Equal(t, ErrNonceMismatch, err)
	assert.Empty(t, token.Value)
	assert.Empty(t, token.RefreshToken)
	// Verify no user was created
	user, _ := userStore.UserByUsername(testutil.Email)
	assert.Empty(t, user.Username)
}

func TestOidcService_OnConfigChanged(t *testing.T) {
	// Setup
	mockServer := testutil.NewOidcTestServerWithToken(t)
	defer mockServer.Close()

	userStore, _ := storage.NewUsersStorage(testutil.NewMemoryStorage(t))
	sessionStore, _ := storage.NewSessionStorage(testutil.NewMemoryStorage(t))
	tokenHolder := storage.NewTokenHolder()
	authStore, _ := storage.NewAuthStorage(userStore, sessionStore, tokenHolder)

	oldConfig := models.OidcConfig{
		IssuerURL: "http://new-issuer-url.com",
		ClientID:  "new-client-id",
	}

	oidcService := NewOidcService(oldConfig, authStore)

	newConfig := models.OidcConfig{
		IssuerURL: mockServer.IssuerURL,
		ClientID:  testutil.ClientID,
	}

	oidcService.OnConfigChanged(newConfig)

	// Verify the config was updated
	redirectURL := "http://localhost:8080/callback"
	state := "teststate"
	nonce := "testnonce"

	authURL, err := oidcService.GetAuthURL(redirectURL, state, nonce)

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, authURL, newConfig.IssuerURL)
	assert.Contains(t, authURL, "client_id="+newConfig.ClientID)
}
