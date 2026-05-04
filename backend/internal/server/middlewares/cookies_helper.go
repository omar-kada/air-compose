// Package middlewares provides HTTP middleware functionality.
package middlewares

import (
	"errors"
	"net/http"
	"time"

	"omar-kada/air-compose/internal/users"
	"omar-kada/air-compose/models"
)

const (
	_tokenKey        = "token"
	_refreshTokenKey = "refreshToken"
	_state           = "state"
	_nonce           = "nonce"
	_originalURL     = "originalURL"
)

func getUsernameFromCookies(r *http.Request, authService users.AuthService) (string, error) {
	token := getTokenFromCookies(r)
	if token.Value == "" {
		return "", errors.New("no auth available")
	}
	return authService.GetUsernameByToken(token)
}

func getTokenFromCookies(r *http.Request) models.Token {
	cookie, err := r.Cookie(_tokenKey)
	if err != nil {
		cookie = &http.Cookie{
			Value: "",
		}
	}
	refreshCookie, err := r.Cookie(_refreshTokenKey)
	if err != nil {
		refreshCookie = &http.Cookie{
			Value: "",
		}
	}
	return models.Token{
		Value:        models.TokenValue(cookie.Value),
		RefreshToken: models.TokenValue(refreshCookie.Value),
	}
}

func setTokenInCookies(w http.ResponseWriter, token models.Token, secureToken bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     _tokenKey,
		Value:    string(token.Value),
		MaxAge:   int(time.Until(token.Expires).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secureToken,
		Path:     "/api",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     _refreshTokenKey,
		Value:    string(token.RefreshToken),
		MaxAge:   int(time.Until(token.RefreshExpires).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secureToken,
		Path:     "/api",
	})
}

func setStateInCookies(w http.ResponseWriter, state, nonce string, secureToken bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     _state,
		Value:    state,
		MaxAge:   int(time.Hour.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secureToken,
		Path:     "/api",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     _nonce,
		Value:    nonce,
		MaxAge:   int(time.Hour.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secureToken,
		Path:     "/api",
	})
}

func getStateFromCookies(r *http.Request) string {
	cookie, err := r.Cookie(_state)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func getNonceFromCookies(r *http.Request) string {
	cookie, err := r.Cookie(_nonce)
	if err != nil {
		return ""
	}
	return cookie.Value
}
