package users

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"omar-kada/air-compose/models"

	"golang.org/x/crypto/bcrypt"
)

const (
	tokenExpiryDuration        = 30 * time.Minute
	refreshTokenExpiryDuration = 30 * 24 * time.Hour
)

// Helper functions
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken() (models.Token, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return models.Token{}, err
	}
	refreshToken := make([]byte, 32)
	_, err = rand.Read(refreshToken)
	if err != nil {
		return models.Token{}, err
	}
	return models.Token{
		Value:          models.TokenValue(hex.EncodeToString(token)),
		Expires:        time.Now().Add(tokenExpiryDuration),
		RefreshToken:   models.TokenValue(hex.EncodeToString(refreshToken)),
		RefreshExpires: time.Now().Add(refreshTokenExpiryDuration),
	}, nil
}
