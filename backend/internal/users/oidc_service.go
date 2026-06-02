// Package users provides user management and authentication services.
package users

import (
	"context"
	"errors"
	"log/slog"
	"omar-kada/air-compose/internal/models"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	// ErrMissingToken indicates that the OAuth token is missing in the response.
	ErrMissingToken = errors.New("missing token in oauth code")
	// ErrNonceMismatch indicates that the nonce in the ID token doesn't match the expected nonce
	ErrNonceMismatch = errors.New("nonce mismatch")
)

// OidcService abstracts oidc operations
type OidcService interface {
	GetAuthURL(redirectURL string, state string, nonce string) (string, error)
	LoginOidc(oauthToken string, nonce string, callbackURL string) (models.Token, error)
	OnConfigChanged(newConfig models.OidcConfig)
}

type oidcService struct {
	config    models.OidcConfig
	authStore AuthStore
}

// NewOidcService creates a new OidcService
func NewOidcService(config models.OidcConfig, authStore AuthStore) OidcService {
	return &oidcService{
		config:    config,
		authStore: authStore,
	}
}

func (s *oidcService) OnConfigChanged(newConfig models.OidcConfig) {
	s.config = newConfig
}

func (s *oidcService) GetAuthURL(redirectURL string, state string, nonce string) (string, error) {
	provider, err := oidc.NewProvider(context.Background(), s.config.IssuerURL)
	if err != nil {
		return "", err
	}
	oauth2Config := s.getConfig(provider, redirectURL)
	return oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce)), nil
}

func (s *oidcService) getConfig(provider *oidc.Provider, redirectURL string) oauth2.Config {
	return oauth2.Config{
		ClientID:     s.config.ClientID,
		ClientSecret: s.config.ClientSecret,
		RedirectURL:  redirectURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "email"},
	}
}

func (s *oidcService) extractUsername(code, nonce, redirectURL string) (string, error) {
	provider, err := oidc.NewProvider(context.Background(), s.config.IssuerURL)
	if err != nil {
		return "", err
	}

	// Configure an OpenID Connect aware OAuth2 client.
	oauth2Config := s.getConfig(provider, redirectURL)

	var verifier = provider.Verifier(&oidc.Config{ClientID: s.config.ClientID})

	oauth2Token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		return "", err
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return "", ErrMissingToken
	}

	// Parse and verify ID Token payload.
	idToken, err := verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		return "", err
	}

	if idToken.Nonce != nonce {
		return "", ErrNonceMismatch
	}

	// Extract custom claims
	var claims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", err
	}
	return claims.Email, nil
}

// LoginOidc authenticates a user and returns their oidc token.
func (s *oidcService) LoginOidc(oauthToken, nonce, callbackURL string) (models.Token, error) {
	// verify token, and then upsert user info + new session
	username, err := s.extractUsername(oauthToken, nonce, callbackURL)
	if err != nil {
		slog.Error("error while authenticating using oidc", "err", err, "oauthToken", oauthToken, "nonce", nonce)
		return models.Token{}, err
	}

	s.authStore.UpsertUser(models.User{
		Username:       username,
		HashedPassword: "",
		Type:           models.UserTypeOIDC,
	})
	return s.authStore.NewAuth(username, generateToken())
}
