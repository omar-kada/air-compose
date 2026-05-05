package models

import "time"

// TokenValue represents a unique authentication token value
type TokenValue string

// Token represents an authentication token with an expiration time
type Token struct {
	Value          TokenValue
	Expires        time.Time
	RefreshToken   TokenValue
	RefreshExpires time.Time
}

// Session represents a user session with authentication details
type Session struct {
	SessionID      uint64 `gorm:"primaryKey;autoGenerate"`
	RefreshToken   string `gorm:"index"`
	RefreshExpires time.Time
	Revoked        bool
	Username       string `gorm:"not null"`
}

// UserType represents the type of user authentication
type UserType string

const (
	// UserTypeLocal represents a locally authenticated user
	UserTypeLocal UserType = "local"
	// UserTypeOIDC represents an OpenID Connect authenticated user
	UserTypeOIDC UserType = "oidc"
)

// User represents a user with credentials and authentication details
type User struct {
	Username       string `gorm:"primaryKey"`
	HashedPassword string
	Type           UserType
}

// Credentials represents user login credentials
type Credentials struct {
	Username string
	Password string
}
