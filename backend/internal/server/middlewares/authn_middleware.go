// Package middlewares provides HTTP middleware functionality.
package middlewares

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/users"
	"omar-kada/air-compose/models"
)

type contextKey string

const (
	_usernameKey contextKey = "username"
)

// ContextWithUsername adds user information to the context.
// @param ctx context.Context - the context to add user information to
// @param user models.User - the user information to add
// @return context.Context - the context with user information added
func ContextWithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, _usernameKey, username)
}

// UsernameFromContext retrieves user information from the context.
// @param ctx context.Context - the context to retrieve user information from
// @return models.User - the user information retrieved
// @return bool - true if user information was found, false otherwise
func UsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(_usernameKey).(string)
	return username, ok
}

// AuthnMiddleware provides authentication middleware.
// @param next http.Handler - the next handler in the chain
// @param authService user.AuthService - the authentication service
// @return http.Handler - the authentication middleware
func AuthnMiddleware(next http.Handler, authService users.AuthService, secureToken bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint, ok := strings.CutPrefix(r.URL.Path, "/api/")
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		switch endpoint {
		case "auth/register":
			if r.Method == http.MethodGet {
				next.ServeHTTP(w, r)
			} else {
				registerHandler(w, r, authService, secureToken)
			}
			return
		case "auth/login":
			loginHandler(w, r, authService, secureToken)
			return
		case "auth/logout":
			logoutHandler(w, r, authService, secureToken)
			return
		case "auth/refresh":
			refreshHandler(w, r, authService, secureToken)
			return
		}
		inWhiteList := isWhitelisted(endpoint, r.Method)

		username, err := getUsernameFromCookies(r, authService)
		if err != nil {
			slog.Error(err.Error())
			if !inWhiteList {
				sendError(w, api.ErrorCodeINVALIDTOKEN)
				return
			}
		}
		r = r.WithContext(ContextWithUsername(r.Context(), username))

		setOriginURLInCookies(w, r.Referer(), secureToken)

		next.ServeHTTP(w, r)
	})
}

var _whitelisted = map[string][]string{
	"user": {"GET"},
}

func isWhitelisted(url, method string) bool {
	if methods, ok := _whitelisted[url]; ok {
		return slices.Contains(methods, method)
	}
	return false
}

func registerHandler(w http.ResponseWriter, r *http.Request, authService users.AuthService, secureToken bool) {
	if r.Method != http.MethodPost {
		sendError(w, api.ErrorCodeNOTALLOWED)
		return
	}
	var req api.Credentials

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		sendErrorMessage(w, api.ErrorCodeINVALIDREQUEST, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		sendErrorMessage(w, api.ErrorCodeINVALIDREQUEST, "Username and password are required")
		return
	}

	token, err := authService.Register(models.Credentials{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		slog.Error(err.Error())
		sendErrorMessage(w, api.ErrorCodeSERVERERROR, "Registration failed")

		return
	}

	setTokenInCookies(w, token, secureToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.BooleanResponse{
		Success: true,
	})
	return
}

func loginHandler(w http.ResponseWriter, r *http.Request, authService users.AuthService, secureToken bool) {
	if r.Method != http.MethodPost {
		sendError(w, api.ErrorCodeNOTALLOWED)
		return
	}

	var req api.Credentials

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		sendErrorMessage(w, api.ErrorCodeINVALIDREQUEST, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		sendErrorMessage(w, api.ErrorCodeINVALIDREQUEST, "Username and password are required")
		return
	}

	auth, err := authService.Login(models.Credentials{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		slog.Error(err.Error())
		sendError(w, api.ErrorCodeINVALIDCREDENTIALS)
		return
	}

	setTokenInCookies(w, auth, secureToken)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.BooleanResponse{
		Success: true,
	})
}

func refreshHandler(w http.ResponseWriter, r *http.Request, authService users.AuthService, secureToken bool) {
	if r.Method != http.MethodPost {
		sendError(w, api.ErrorCodeNOTALLOWED)
		return
	}

	token := getTokenFromCookies(r)

	if token.RefreshToken == "" {
		slog.Error("invalid refresh token value")
		sendError(w, api.ErrorCodeINVALIDCREDENTIALS)
		return
	}

	newToken, err := authService.RefreshToken(token)
	if err != nil {
		slog.Error(err.Error())
		sendError(w, api.ErrorCodeINVALIDCREDENTIALS)
		return
	}

	setTokenInCookies(w, newToken, secureToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.BooleanResponse{
		Success: true,
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request, authService users.AuthService, secureToken bool) {
	if r.Method != http.MethodPost {
		sendError(w, api.ErrorCodeNOTALLOWED)
		return
	}

	token := getTokenFromCookies(r)
	if token.Value == "" || token.RefreshToken == "" {
		slog.Error("invalid token value")
		sendError(w, api.ErrorCodeINVALIDTOKEN)
		return
	}

	err := authService.Logout(token)
	if err != nil {
		slog.Error(err.Error())
		sendError(w, api.ErrorCodeINVALIDTOKEN)
		return
	}

	setTokenInCookies(w, models.Token{
		Value:          "",
		Expires:        time.Unix(0, 0),
		RefreshToken:   "",
		RefreshExpires: time.Unix(0, 0),
	}, secureToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.BooleanResponse{
		Success: true,
	})
}
