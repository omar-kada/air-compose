package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/internal/users"
	"omar-kada/air-compose/models"
	"omar-kada/air-compose/testutil"

	"github.com/stretchr/testify/assert"
)

var userCreds = models.Credentials{
	Username: "username",
	Password: "password",
}

func newUsersService(t *testing.T) users.Service {
	userStore, err := storage.NewUsersStorage(testutil.NewMemoryStorage(t))
	assert.NoError(t, err)
	SessionStore, err := storage.NewSessionStorage(testutil.NewMemoryStorage(t))
	assert.NoError(t, err)

	tokenHolder := storage.NewTokenHolder()
	store, err := storage.NewAuthStorage(userStore, SessionStore, tokenHolder)
	assert.NoError(t, err)

	return users.NewService(store)
}

func withInitUsers(t *testing.T, userService users.Service, creds models.Credentials) (users.Service, models.Token) {
	token, err := userService.Register(creds)
	assert.NoError(t, err)

	return userService, token
}

func TestAuthnMiddleware_Register(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"testuser","password":"testpass"}`
	req := httptest.NewRequest("POST", "https://example.com/api/auth/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assertTokenCookiesAreNot(t, rr, "", "")
	testutil.AssertCookiesAreSecure(t, rr, _tokenKey, _refreshTokenKey)
}

func TestAuthnMiddleware_RegisterGet(t *testing.T) {
	userService, _ := withInitUsers(t, newUsersService(t), userCreds)

	// The GET /api/auth/register should pass through to the next handler
	called := false
	handler := AuthnMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}), userService)

	req := httptest.NewRequest("GET", "https://example.com/api/auth/register", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.True(t, called, "next handler should be called for GET /api/auth/register")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"ok":true}`, rr.Body.String())
	assertTokenCookiesNotExisting(t, rr)
	testutil.AssertCookiesAreSecure(t, rr, _tokenKey, _refreshTokenKey)
}

func TestAuthnMiddleware_Login(t *testing.T) {
	userService := newUsersService(t)
	userService, token := withInitUsers(t, userService, userCreds)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"username","password":"password"}`
	req := httptest.NewRequest("POST", "https://example.com/api/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assertTokenCookiesAreNot(t, rr, string(token.Value), string(token.RefreshToken))
	assertTokenCookiesAreNot(t, rr, "", "")
	testutil.AssertCookiesAreSecure(t, rr, _tokenKey, _refreshTokenKey)
}

func TestAuthnMiddleware_Logout(t *testing.T) {
	userService := newUsersService(t)
	userService, token := withInitUsers(t, userService, userCreds)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("POST", "https://example.com/api/auth/logout", http.NoBody)
	req.AddCookie(&http.Cookie{
		Name:    _tokenKey,
		Value:   string(token.Value),
		Expires: token.Expires,
	})
	req.AddCookie(&http.Cookie{
		Name:    _refreshTokenKey,
		Value:   string(token.RefreshToken),
		Expires: token.RefreshExpires,
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assertTokenCookiesAre(t, rr, "", "")
	testutil.AssertCookiesAreSecure(t, rr, _tokenKey, _refreshTokenKey)
}

func TestAuthnMiddleware_AuthorizedAccess(t *testing.T) {
	userService := newUsersService(t)
	userService, token := withInitUsers(t, userService, userCreds)

	handler := AuthnMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, ok := UsernameFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, "username", username)
		w.WriteHeader(http.StatusOK)
	}), userService)

	req := httptest.NewRequest("GET", "https://example.com/api/auth/protected", http.NoBody)
	req.AddCookie(&http.Cookie{
		Name:  _tokenKey,
		Value: string(token.Value),
	})
	req.AddCookie(&http.Cookie{
		Name:  _refreshTokenKey,
		Value: string(token.RefreshToken),
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthnMiddleware_WhitelistedAccess(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), userService)

	req := httptest.NewRequest("GET", "https://example.com/api/user", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_RegisterInvalidRequestBody(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"testuser"}` // missing password
	req := httptest.NewRequest("POST", "https://example.com/api/auth/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_RegisterMissingCredentials(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"","password":""}` // empty username and password
	req := httptest.NewRequest("POST", "https://example.com/api/auth/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_RegisterFailure(t *testing.T) {
	userService := newUsersService(t)
	// Register first user to prevent registration
	_, _ = withInitUsers(t, userService, userCreds)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"testuser","password":"testpass"}`
	req := httptest.NewRequest("POST", "https://example.com/api/auth/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_LoginInvalidMethod(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("GET", "https://example.com/api/auth/login", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_LoginInvalidRequestBody(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"testuser"}` // missing password
	req := httptest.NewRequest("POST", "https://example.com/api/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_LoginMissingCredentials(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"","password":""}` // empty username and password
	req := httptest.NewRequest("POST", "https://example.com/api/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_LoginFailure(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"wronguser","password":"wrongpass"}`
	req := httptest.NewRequest("POST", "https://example.com/api/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_LogoutInvalidMethod(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("GET", "https://example.com/api/auth/logout", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_LogoutMissingToken(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("POST", "https://example.com/api/auth/logout", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_LogoutFailure(t *testing.T) {
	userService := newUsersService(t)
	userService, token := withInitUsers(t, userService, userCreds)

	// Invalidate the token by modifying it
	invalidToken := token
	invalidToken.Value = "invalidtoken"

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("POST", "https://example.com/api/auth/logout", http.NoBody)
	req.AddCookie(&http.Cookie{
		Name:  _tokenKey,
		Value: string(invalidToken.Value),
	})
	req.AddCookie(&http.Cookie{
		Name:  _refreshTokenKey,
		Value: string(invalidToken.RefreshToken),
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthnMiddleware_Refresh(t *testing.T) {
	userService := newUsersService(t)
	userService, token := withInitUsers(t, userService, userCreds)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("POST", "https://example.com/api/auth/refresh", http.NoBody)
	req.AddCookie(&http.Cookie{
		Name:  _refreshTokenKey,
		Value: string(token.RefreshToken),
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assertTokenCookiesAreNot(t, rr, string(token.Value), string(token.RefreshToken))
	assertTokenCookiesAreNot(t, rr, "", "")
	testutil.AssertCookiesAreSecure(t, rr, _tokenKey, _refreshTokenKey)
}

func TestAuthnMiddleware_RefreshInvalidMethod(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("GET", "https://example.com/api/auth/refresh", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_RefreshMissingToken(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("POST", "https://example.com/api/auth/refresh", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}

func TestAuthnMiddleware_RefreshFailure(t *testing.T) {
	userService := newUsersService(t)
	userService, token := withInitUsers(t, userService, userCreds)

	// Invalidate the refresh token by modifying it
	invalidToken := token
	invalidToken.RefreshToken = "invalidtoken"

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("POST", "https://example.com/api/auth/refresh", http.NoBody)
	req.AddCookie(&http.Cookie{
		Name:  _refreshTokenKey,
		Value: string(invalidToken.RefreshToken),
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assertTokenCookiesNotExisting(t, rr)
}
func TestAuthnMiddleware_InsecureCookies(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// Should not be called
		t.Fail()
	}), userService)

	reqBody := `{"username":"testuser","password":"testpass"}`
	req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	cookieMap := testutil.CookiesToMap(rr.Result().Cookies())
	assert.NotEqual(t, "", cookieMap[_tokenKey].Value)
	assert.NotEqual(t, "", cookieMap[_refreshTokenKey].Value)
	assertTokenCookiesAreNot(t, rr, "", "")
	testutil.AssertCookiesAreNotSecure(t, rr, _tokenKey, _refreshTokenKey)
}

func assertTokenCookiesAreNot(t *testing.T, rr *httptest.ResponseRecorder, expectedToken, expectedRefreshToken string) {
	cookieMap := testutil.CookiesToMap(rr.Result().Cookies())
	assert.Contains(t, cookieMap, _tokenKey, "Token cookie should be set")
	assert.NotEqual(t, expectedToken, cookieMap[_tokenKey].Value)
	assert.Contains(t, cookieMap, _refreshTokenKey, "Token cookie should be set")
	assert.NotEqual(t, expectedRefreshToken, cookieMap[_refreshTokenKey].Value)
}

func assertTokenCookiesAre(t *testing.T, rr *httptest.ResponseRecorder, expectedToken, expectedRefreshToken string) {

	cookieMap := testutil.CookiesToMap(rr.Result().Cookies())
	assert.Contains(t, cookieMap, _tokenKey, "Token cookie should be set")
	assert.Equal(t, expectedToken, cookieMap[_tokenKey].Value)
	assert.Contains(t, cookieMap, _refreshTokenKey, "Token cookie should be set")
	assert.Equal(t, expectedRefreshToken, cookieMap[_refreshTokenKey].Value)
}

func assertTokenCookiesNotExisting(t *testing.T, rr *httptest.ResponseRecorder) {

	cookieMap := testutil.CookiesToMap(rr.Result().Cookies())
	assert.NotContains(t, cookieMap, _tokenKey, "Token cookie should be set")
	assert.NotContains(t, cookieMap, _refreshTokenKey, "Token cookie should be set")
}
