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
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
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
	req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	checkCookiesAreNot(t, rr, "", "")
}

func TestAuthnMiddleware_RegisterGet(t *testing.T) {
	userService, _ := withInitUsers(t, newUsersService(t), userCreds)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("GET", "/api/auth/register", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"registered": true}`, rr.Body.String())
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_Login(t *testing.T) {
	userService := newUsersService(t)
	userService, token := withInitUsers(t, userService, userCreds)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"username","password":"password"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	checkCookiesAreNot(t, rr, string(token.Value), string(token.RefreshToken))
	checkCookiesAreNot(t, rr, "", "")
}

func TestAuthnMiddleware_Logout(t *testing.T) {
	userService := newUsersService(t)
	userService, token := withInitUsers(t, userService, userCreds)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("POST", "/api/auth/logout", http.NoBody)
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
	checkCookiesAre(t, rr, "", "")
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

	req := httptest.NewRequest("GET", "/api/auth/protected", http.NoBody)
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

	req := httptest.NewRequest("GET", "/api/user", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_RegisterInvalidRequestBody(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"testuser"}` // missing password
	req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_RegisterMissingCredentials(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"","password":""}` // empty username and password
	req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_RegisterFailure(t *testing.T) {
	userService := newUsersService(t)
	// Register first user to prevent registration
	_, _ = withInitUsers(t, userService, userCreds)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"testuser","password":"testpass"}`
	req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_LoginInvalidMethod(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("GET", "/api/auth/login", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_LoginInvalidRequestBody(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"testuser"}` // missing password
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_LoginMissingCredentials(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"","password":""}` // empty username and password
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_LoginFailure(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	reqBody := `{"username":"wronguser","password":"wrongpass"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_LogoutInvalidMethod(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("GET", "/api/auth/logout", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_LogoutMissingToken(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("POST", "/api/auth/logout", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	checkCookiesAre(t, rr, "", "")
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

	req := httptest.NewRequest("POST", "/api/auth/logout", http.NoBody)
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

	req := httptest.NewRequest("POST", "/api/auth/refresh", http.NoBody)
	req.AddCookie(&http.Cookie{
		Name:  _refreshTokenKey,
		Value: string(token.RefreshToken),
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	checkCookiesAreNot(t, rr, string(token.Value), string(token.RefreshToken))
	checkCookiesAreNot(t, rr, "", "")
}

func TestAuthnMiddleware_RefreshInvalidMethod(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("GET", "/api/auth/refresh", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func TestAuthnMiddleware_RefreshMissingToken(t *testing.T) {
	userService := newUsersService(t)

	handler := AuthnMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), userService)

	req := httptest.NewRequest("POST", "/api/auth/refresh", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	checkCookiesAre(t, rr, "", "")
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

	req := httptest.NewRequest("POST", "/api/auth/refresh", http.NoBody)
	req.AddCookie(&http.Cookie{
		Name:  _refreshTokenKey,
		Value: string(invalidToken.RefreshToken),
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	checkCookiesAre(t, rr, "", "")
}

func checkCookiesAreNot(t *testing.T, rr *httptest.ResponseRecorder, expectedToken, expectedRefreshToken string) {
	cookies := rr.Result().Cookies()
	var tokenCookie, refreshTokenCookie http.Cookie

	for _, cookie := range cookies {
		if cookie.Name == _tokenKey {
			tokenCookie = *cookie
		}
		if cookie.Name == _refreshTokenKey {
			refreshTokenCookie = *cookie
		}
	}

	assert.NotNil(t, tokenCookie, "Token cookie should be set")
	assert.NotEqual(t, expectedToken, tokenCookie.Value, "Token cookie value should match")

	assert.NotNil(t, refreshTokenCookie, "Refresh token cookie should be set")
	assert.NotEqual(t, expectedRefreshToken, refreshTokenCookie.Value, "Refresh token cookie value should match")
}

func checkCookiesAre(t *testing.T, rr *httptest.ResponseRecorder, expectedToken, expectedRefreshToken string) {
	cookies := rr.Result().Cookies()
	var tokenCookie, refreshTokenCookie http.Cookie

	for _, cookie := range cookies {
		if cookie.Name == _tokenKey {
			tokenCookie = *cookie
		}
		if cookie.Name == _refreshTokenKey {
			refreshTokenCookie = *cookie
		}
	}

	assert.NotNil(t, tokenCookie, "Token cookie should be set")
	assert.Equal(t, expectedToken, tokenCookie.Value, "Token cookie value should match")

	assert.NotNil(t, refreshTokenCookie, "Refresh token cookie should be set")
	assert.Equal(t, expectedRefreshToken, refreshTokenCookie.Value, "Refresh token cookie value should match")
}
