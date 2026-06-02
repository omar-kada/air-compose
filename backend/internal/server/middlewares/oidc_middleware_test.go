package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/users"
	"omar-kada/air-compose/testutil"

	"github.com/stretchr/testify/assert"
)

func newOidcService(t *testing.T) (*testutil.MockOIDCServer, users.OidcService) {
	server := testutil.NewOidcTestServerWithToken(t)

	userStore, err := users.NewUsersStorage(testutil.NewMemoryStorage(t))
	assert.NoError(t, err)
	SessionStore, err := users.NewSessionStorage(testutil.NewMemoryStorage(t))
	assert.NoError(t, err)

	tokenHolder := users.NewTokenHolder()
	store, err := users.NewAuthStorage(userStore, SessionStore, tokenHolder)
	assert.NoError(t, err)

	return server, users.NewOidcService(models.OidcConfig{
		IssuerURL: server.IssuerURL,
		ClientID:  testutil.ClientID,
	}, store)
}

func TestOidcMiddleware_LoginRedirect(t *testing.T) {
	server, oidcService := newOidcService(t)
	defer server.Close()

	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), oidcService)

	req := httptest.NewRequest("GET", "/api/oidc/login", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "/auth?")

	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == _state || cookie.Name == _nonce {
			assert.NotEmpty(t, cookie.Value, "Cookie should be set")
			assert.True(t, cookie.HttpOnly, "Cookie should be HttpOnly")
			assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite, "Cookie should be SameSiteStrictMode")
		}
	}
}

func TestOidcMiddleware_LoginWrongConfig(t *testing.T) {
	server, oidcService := newOidcService(t)
	defer server.Close()

	oidcService.OnConfigChanged(models.OidcConfig{
		IssuerURL: "http://invalid-url.com",
		ClientID:  "invalid-client-id",
	})

	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), oidcService)

	req := httptest.NewRequest("GET", "/api/oidc/login", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == _state || cookie.Name == _nonce {
			assert.Fail(t, "cookie shouldn't be present on error", cookie.Name)
		}
	}
}

func TestOidcMiddleware_LoginInvalidMethod(t *testing.T) {
	_, oidcService := newOidcService(t)
	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), oidcService)

	req := httptest.NewRequest("POST", "/api/oidc/login", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestOidcMiddleware_Callback(t *testing.T) {
	server, oidcService := newOidcService(t)
	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), oidcService)

	code := server.SignIDToken(testutil.ClientID, "user", map[string]any{
		"email": "test@example.com",
	})

	req := httptest.NewRequest("GET", "/api/oidc/callback?code="+code+"&state=teststate", http.NoBody)
	rr := httptest.NewRecorder()

	req.AddCookie(&http.Cookie{
		Name:  _state,
		Value: "teststate",
	})

	req.AddCookie(&http.Cookie{
		Name:  _nonce,
		Value: testutil.Nonce,
	})
	req.AddCookie(&http.Cookie{
		Name:  _originURL,
		Value: "http://test.com",
	})

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "http://test.com")

	// Check that state and nonce cookies are cleared
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == _state || cookie.Name == _nonce {
			assert.Empty(t, cookie.Value, "State and nonce cookies should be cleared after login")
		}
	}
}

func TestOidcMiddleware_InvalidMethod(t *testing.T) {
	_, oidcService := newOidcService(t)
	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), oidcService)

	req := httptest.NewRequest("POST", "/api/oidc/callback", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestOidcMiddleware_MissingCode(t *testing.T) {
	_, oidcService := newOidcService(t)
	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), oidcService)

	req := httptest.NewRequest("GET", "/api/oidc/callback?state=teststate", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOidcMiddleware_InvalidState(t *testing.T) {
	_, oidcService := newOidcService(t)
	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), oidcService)

	req := httptest.NewRequest("GET", "/api/oidc/callback?code=testcode&state=invalidstate", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestOidcMiddleware_OidcLoginFailure(t *testing.T) {
	_, oidcService := newOidcService(t)
	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), oidcService)

	req := httptest.NewRequest("GET", "/api/oidc/callback?code=invalidcode&state=teststate", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestOidcMiddleware_InsecureCookies(t *testing.T) {
	server, oidcService := newOidcService(t)
	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fail() // shouldn't be called
	}), oidcService)

	code := server.SignIDToken(testutil.ClientID, "user", map[string]any{
		"email": "test@example.com",
	})

	req := httptest.NewRequest("GET", "/api/oidc/callback?code="+code+"&state=teststate", http.NoBody)
	req.TLS = nil
	rr := httptest.NewRecorder()

	req.AddCookie(&http.Cookie{
		Name:  _state,
		Value: "teststate",
	})
	req.AddCookie(&http.Cookie{
		Name:  _nonce,
		Value: testutil.Nonce,
	})

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	cookies := rr.Result().Cookies()

	for _, cookie := range cookies {
		if cookie.Name == _tokenKey {
			assert.False(t, cookie.Secure, "Token cookie should not be secure")
			assert.True(t, cookie.HttpOnly, "Token cookie should be HttpOnly")
			assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite, "Token cookie should be SameSiteStrictMode")
		}
		if cookie.Name == _refreshTokenKey {
			assert.False(t, cookie.Secure, "Token cookie should not be secure")
			assert.True(t, cookie.HttpOnly, "Refresh token cookie should be HttpOnly")
			assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite, "Refresh token cookie should be SameSiteStrictMode")
		}
	}
}

func TestOidcMiddleware_NextHandlerCalled(t *testing.T) {
	_, oidcService := newOidcService(t)
	called := false

	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}), oidcService)

	req := httptest.NewRequest("GET", "/api/some-other-endpoint", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.True(t, called, "Next handler should be called for non-OIDC endpoints")
}

func TestOidcMiddleware_OriginURLSet(t *testing.T) {
	_, oidcService := newOidcService(t)
	called := false

	handler := OidcMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}), oidcService)

	req := httptest.NewRequest("GET", "/api/some-other-endpoint", http.NoBody)
	req.Header.Set("Referer", "http://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.True(t, called, "Next handler should be called")
	cookiesMap := testutil.CookiesToMap(rr.Result().Cookies())
	cookie, found := cookiesMap[_originURL]
	assert.True(t, found, "Origin URL cookie should be set")
	assert.Equal(t, "http://example.com", cookie.Value, "Origin URL should be set from Referer header")
}
