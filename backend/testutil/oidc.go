package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/coreos/go-oidc/v3/oidc/oidctest"
)

// ClientID is the default client ID used for testing OIDC flows.
const ClientID = "test-client"

// User is the default user ID used for testing OIDC flows.
const User = "test-user"

// Email is the default email address used for testing OIDC flows.
const Email = "test-user@example.com"

// NewOidcTestServerWithToken starts an OIDC test server with a /token endpoint for OAuth2 flows.
// Returns a models.OidcConfig and a cleanup function to close the servers.
func NewOidcTestServerWithToken(t *testing.T) *MockOIDCServer {
	t.Helper()

	// Generate a signing key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	keyID := "test-key-1"
	algorithm := oidc.RS256

	// Create the test server with a public key
	s := &oidctest.Server{
		PublicKeys: []oidctest.PublicKey{{
			PublicKey: priv.Public(),
			KeyID:     keyID,
			Algorithm: algorithm,
		}},
	}

	mockServer := &MockOIDCServer{
		PrivateKey: priv,
		KeyID:      keyID,
		Algorithm:  algorithm,
		oidcServer: s,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		resp := map[string]interface{}{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"id_token": mockServer.SignIDToken(ClientID, User, map[string]any{
				"email": Email,
			}),
			"refresh_token": "test-refresh-token",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	// Proxy all other requests to the oidctest server
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.ServeHTTP(w, r)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(func() { srv.Close() })

	s.SetIssuer(srv.URL) // Must be called before serving
	mockServer.IssuerURL = srv.URL
	mockServer.Server = srv

	return mockServer
}

// MockOIDCServer provides a mock issuer for unit tests, strictly using the oidctest library.
// It does NOT include a functional /token endpoint and is NOT designed to validate client secrets.
type MockOIDCServer struct {
	Server     *httptest.Server
	IssuerURL  string
	PrivateKey *rsa.PrivateKey
	KeyID      string
	Algorithm  string
	oidcServer *oidctest.Server
}

// StartMockOIDCServer creates a mock OIDC provider that handles discovery and key serving only.
// It does NOT support client authentication or token exchange. Use this to test your client's
// core verification logic with custom test tokens.
func StartMockOIDCServer(t *testing.T) *MockOIDCServer {
	t.Helper()

	// Generate a signing key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	keyID := "test-key-1"
	algorithm := oidc.RS256

	// Create the test server with a public key
	s := &oidctest.Server{
		PublicKeys: []oidctest.PublicKey{{
			PublicKey: priv.Public(),
			KeyID:     keyID,
			Algorithm: algorithm,
		}},
	}

	httpSrv := httptest.NewServer(s)
	t.Cleanup(func() { httpSrv.Close() })

	s.SetIssuer(httpSrv.URL) // Must be called before serving

	return &MockOIDCServer{
		Server:     httpSrv,
		IssuerURL:  httpSrv.URL,
		PrivateKey: priv,
		KeyID:      keyID,
		Algorithm:  algorithm,
		oidcServer: s,
	}
}

// SignIDToken signs an ID token for the given user and audiences.
func (m *MockOIDCServer) SignIDToken(clientID, userID string, customClaims map[string]interface{}) string {
	// Build the required minimal claims per OIDC spec
	claims := map[string]interface{}{
		"iss": m.IssuerURL,
		"aud": clientID,
		"sub": userID,
		"exp": 360000000000,
		// Add other claims like exp, iat, etc. if needed
	}
	for k, v := range customClaims {
		claims[k] = v
	}

	// The library expects a JSON string
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		panic(err)
	}

	// Sign and return the token as a string
	return oidctest.SignIDToken(m.PrivateKey, m.KeyID, m.Algorithm, string(claimsJSON))
}

// Close closes the underlying HTTP server.
func (m *MockOIDCServer) Close() {
	m.Server.Close()
}
