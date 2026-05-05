// Package middlewares provides HTTP middleware functionality.
package middlewares

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/users"
)

const (
	codeParam        = "code"
	stateParam       = "state"
	callbackEndpoint = "/api/oidc/callback"
)

// OidcMiddleware provides oidc middleware.
// @param next http.Handler - the next handler in the chain
// @param oidcService user.OidcService - the oidc service
// @return http.Handler - the authentication middleware
func OidcMiddleware(next http.Handler, oidcService users.OidcService, secureToken bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/oidc/login":
			oidcLoginRedirectHandler(w, r, oidcService, secureToken)
			return
		case "/api/oidc/callback":
			oidcCallbackHandler(w, r, oidcService, secureToken)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func oidcCallbackHandler(w http.ResponseWriter, r *http.Request, oidcService users.OidcService, secureToken bool) {
	if r.Method != http.MethodGet {
		sendError(w, api.ErrorCodeNOTALLOWED)
		return
	}

	code := r.URL.Query().Get(codeParam)
	if code == "" {
		sendErrorMessage(w, api.ErrorCodeINVALIDREQUEST, "oauth code not found")
		return
	}
	state := r.URL.Query().Get(stateParam)
	expectedState := getStateFromCookies(r)
	if state != expectedState {
		sendErrorMessage(w, api.ErrorCodeINVALIDCREDENTIALS, "oauth state error")
		return
	}

	token, err := oidcService.LoginOidc(code, getNonceFromCookies(r))
	if err != nil {
		slog.Error(err.Error())
		sendErrorMessage(w, api.ErrorCodeSERVERERROR, "OIDC authentication failed")
		return
	}

	setTokenInCookies(w, token, secureToken)
	setStateInCookies(w, "", "", secureToken)
	http.Redirect(w, r, getBaseURL(r), http.StatusFound)
}

func oidcLoginRedirectHandler(w http.ResponseWriter, r *http.Request, oidcService users.OidcService, secureToken bool) {
	if r.Method != http.MethodGet {
		sendError(w, api.ErrorCodeNOTALLOWED)
		return
	}
	state, err := generateState()
	if err != nil {
		slog.Error("error while generating oauth state", "err", err)
		sendError(w, api.ErrorCodeSERVERERROR)
		return
	}
	nonce, err := generateState()
	if err != nil {
		slog.Error("error while generating oauth nonce", "err", err)
		sendError(w, api.ErrorCodeSERVERERROR)
		return
	}

	authURL, err := oidcService.GetAuthURL(getBaseURL(r)+callbackEndpoint, state, nonce)
	if err != nil {
		slog.Error("error while getting auth URL", "err", err)
		sendErrorMessage(w, api.ErrorCodeSERVERERROR, "error while getting auth URL")
		return
	}
	setStateInCookies(w, state, nonce, secureToken)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func generateState() (string, error) {
	randState := make([]byte, 32)
	_, err := rand.Read(randState)
	return base64.RawURLEncoding.EncodeToString(randState), err
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
