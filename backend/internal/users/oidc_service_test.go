package users

import (
	"net/url"
	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupOidcTest(t *testing.T) (*testutil.MockOIDCServer, OidcService, UserStorage) {
	mockServer := testutil.NewOidcTestServerWithToken(t)
	configStore, err := config.NewConfigStore(t.TempDir()+"/config.yaml", events.NewBus(1))
	if err != nil {
		t.Fatal(err)
	}

	userStore, _ := NewUsersStorage(testutil.NewMemoryStorage(t))
	sessionStore, _ := NewSessionStorage(testutil.NewMemoryStorage(t))
	tokenHolder := NewTokenHolder()
	authStore, _ := NewAuthStorage(userStore, sessionStore, tokenHolder)

	configStore.Update(models.Config{
		Settings: models.Settings{
			Oidc: models.OidcConfig{
				IssuerURL: mockServer.IssuerURL,
				ClientID:  testutil.ClientID,
			},
		},
	})
	t.Cleanup(mockServer.Close)
	return mockServer, NewOidcService(configStore, authStore), userStore
}

func TestOidcService_GetAuthURL(t *testing.T) {
	mockServer, oidcService, _ := setupOidcTest(t)

	redirectURL := "http://localhost:8080/callback"
	state := "teststate"
	nonce := "testnonce"

	authURL, err := oidcService.GetAuthURL(redirectURL, state, nonce)

	assert.NoError(t, err)
	assert.Contains(t, authURL, mockServer.IssuerURL)
	assert.Contains(t, authURL, "response_type=code")
	assert.Contains(t, authURL, "client_id="+testutil.ClientID)
	assert.Contains(t, authURL, "redirect_uri="+url.QueryEscape(redirectURL))
	assert.Contains(t, authURL, "state="+state)
	assert.Contains(t, authURL, "nonce="+nonce)
}

func TestOidcService_LoginOidc(t *testing.T) {
	mockServer, oidcService, userStore := setupOidcTest(t)

	code := mockServer.SignIDToken(testutil.ClientID, testutil.User, map[string]any{
		"email":        testutil.Email,
		"redirect-uri": "callback-url",
	})

	token, err := oidcService.LoginOidc(code, testutil.Nonce, "callback-url")

	assert.NoError(t, err)
	assert.NotEmpty(t, token.Value)
	assert.NotEmpty(t, token.RefreshToken)
	user, err := userStore.UserByUsername(testutil.Email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, testutil.Email, user.Username)
}

func TestOidcService_LoginOidc_NonceMismatch(t *testing.T) {
	mockServer, oidcService, userStore := setupOidcTest(t)

	code := mockServer.SignIDToken(testutil.ClientID, testutil.User, map[string]any{
		"email":        testutil.Email,
		"redirect-uri": "callback-url",
	})

	token, err := oidcService.LoginOidc(code, "wrong-nonce", "")

	assert.Error(t, err)
	assert.Equal(t, ErrNonceMismatch, err)
	assert.Empty(t, token.Value)
	assert.Empty(t, token.RefreshToken)
	user, _ := userStore.UserByUsername(testutil.Email)
	assert.Empty(t, user.Username)
}
