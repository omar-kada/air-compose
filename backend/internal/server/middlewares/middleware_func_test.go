package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthnMiddlewareFunc(t *testing.T) {
	userService := newUsersService(t)

	mw := AuthnMiddlewareFunc(userService)
	assert.NotNil(t, mw)

	called := false
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}))

	// Whitelisted endpoint should pass through
	req := httptest.NewRequest("GET", "/api/user", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, called, "next handler should be called for whitelisted endpoint")
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthnMiddlewareFunc_NonWhitelistedWithoutAuth(t *testing.T) {
	userService := newUsersService(t)

	mw := AuthnMiddlewareFunc(userService)
	called := false
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}))

	req := httptest.NewRequest("GET", "/api/config", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.False(t, called, "next handler should not be called without auth")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestOidcMiddlewareFunc(t *testing.T) {
	_, oidcService, _ := newOidcService(t)

	mw := OidcMiddlewareFunc(oidcService)
	assert.NotNil(t, mw)

	called := false
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}))

	// Non-OIDC endpoint should pass through
	req := httptest.NewRequest("GET", "/api/config", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, called, "next handler should be called for non-OIDC endpoints")
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOidcMiddlewareFunc_InvalidEndpoint(t *testing.T) {
	_, oidcService, _ := newOidcService(t)

	mw := OidcMiddlewareFunc(oidcService)
	called := false
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}))

	req := httptest.NewRequest("GET", "/api/oidc/invalid", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.False(t, called, "next handler should not be called for invalid OIDC endpoint")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
