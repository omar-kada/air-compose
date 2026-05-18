// Package middlewares provides HTTP middleware functionality.
package middlewares

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"strings"

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
func OidcMiddleware(next http.Handler, oidcService users.OidcService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oidcOperation, ok := strings.CutPrefix(r.URL.Path, "/api/oidc/")
		if ok {
			switch oidcOperation {
			case "login":
				oidcLoginRedirectHandler(w, r, oidcService)
			case "callback":
				oidcCallbackHandler(w, r, oidcService)
			default:
				sendError(w, api.ErrorCodeINVALIDREQUEST)
			}
			return
		}

		setOriginURLInCookies(w, r.Referer(), isTLS(r))
		next.ServeHTTP(w, r)
	})
}

func oidcCallbackHandler(w http.ResponseWriter, r *http.Request, oidcService users.OidcService) {
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

	token, err := oidcService.LoginOidc(code, getNonceFromCookies(r), getCallbackURL(r))
	if err != nil {
		slog.Error(err.Error())
		sendErrorMessage(w, api.ErrorCodeSERVERERROR, "OIDC authentication failed")
		return
	}

	originURL := getOriginURLFromCookies(r)
	if originURL == "" {
		originURL = getBaseURL(r)
	}
	setTokenInCookies(w, token, isTLS(r))
	setStateInCookies(w, "", "", isTLS(r))

	http.Redirect(w, r, originURL, http.StatusFound)
}

func oidcLoginRedirectHandler(w http.ResponseWriter, r *http.Request, oidcService users.OidcService) {
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

	authURL, err := oidcService.GetAuthURL(getCallbackURL(r), state, nonce)
	if err != nil {
		slog.Error("error while getting auth URL", "err", err)
		sendErrorMessage(w, api.ErrorCodeSERVERERROR, "error while getting auth URL")
		return
	}
	setStateInCookies(w, state, nonce, isTLS(r))
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

func getCallbackURL(r *http.Request) string {
	return getBaseURL(r) + callbackEndpoint
}
