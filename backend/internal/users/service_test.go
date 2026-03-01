package users

import (
	"testing"
	"time"

	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"
	"omar-kada/air-compose/testutil"

	"github.com/stretchr/testify/assert"
)

func TestLogin_Success(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	credentials := models.Credentials{
		Username: "testuser",
		Password: "password",
	}

	hashedPassword, _ := hashPassword(credentials.Password)
	mockUser := models.User{
		Username:       credentials.Username,
		HashedPassword: hashedPassword,
	}

	store.UpsertUser(mockUser)

	token, err := service.Login(credentials)

	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestLogin_UserNotFound(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	credentials := models.Credentials{
		Username: "nonexistentuser",
		Password: "password",
	}

	token, err := service.Login(credentials)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Zero(t, token)
}

func TestLogin_InvalidPassword(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	credentials := models.Credentials{
		Username: "testuser",
		Password: "wrongpassword",
	}

	hashedPassword, _ := hashPassword("correctpassword")
	mockUser := models.User{
		Username:       credentials.Username,
		HashedPassword: hashedPassword,
	}

	store.UpsertUser(mockUser)

	token, err := service.Login(credentials)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPassword)
	assert.Zero(t, token)
}

func TestIsRegistered_HasUsers(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	mockUser := models.User{
		Username: "testuser",
	}

	store.UpsertUser(mockUser)

	isRegistered, err := service.IsRegistered()

	assert.NoError(t, err)
	assert.True(t, isRegistered)
}

func TestIsRegistered_NoUsers(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	isRegistered, err := service.IsRegistered()

	assert.NoError(t, err)
	assert.False(t, isRegistered)
}

func TestRegister_Success(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	credentials := models.Credentials{
		Username: "newuser",
		Password: "password",
	}

	token, err := service.Register(credentials)

	assert.NoError(t, err)
	assert.NotZero(t, token)
}

func TestRegister_AlreadyRegistered(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	credentials := models.Credentials{
		Username: "existinguser",
		Password: "password",
	}

	mockUser := models.User{
		Username: credentials.Username,
	}

	store.UpsertUser(mockUser)

	token, err := service.Register(credentials)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAlreadyRegistered)
	assert.Zero(t, token)
}

func TestLogout_Success(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	pass, _ := hashPassword("password")
	mockUser := models.User{
		Username:       "testuser",
		HashedPassword: pass,
	}

	store.UpsertUser(mockUser)

	token, err := service.Login(models.Credentials{
		Username: "testuser",
		Password: "password",
	})
	assert.NoError(t, err)

	err = service.Logout(token)
	assert.NoError(t, err)

	user, _ := service.GetUsernameByToken(token)
	assert.Zero(t, user)
}

func TestLogout_UserNotFound(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	err = service.Logout(models.Token{Value: "invalidtoken", RefreshToken: "invalidRefreshToken"})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestGetUserByToken_Success(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	pass, _ := hashPassword("password")
	mockUser := models.User{
		Username:       "testuser",
		HashedPassword: pass,
	}

	_, err = store.UpsertUser(mockUser)
	assert.NoError(t, err)

	token, err := service.Login(models.Credentials{
		Username: "testuser",
		Password: "password",
	})
	assert.NoError(t, err)
	user, err := service.GetUsernameByToken(token)

	assert.NoError(t, err)
	assert.Equal(t, mockUser.Username, user)
}

func TestGetUserByToken_UserNotFound(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	user, err := service.GetUsernameByToken(models.Token{Value: "invalidtoken"})

	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Zero(t, user)
}

func TestGetUser_Success(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	username := "testuser"
	mockUser := models.User{
		Username: username,
	}

	store.UpsertUser(mockUser)

	user, err := service.GetUser(username)

	assert.NoError(t, err)
	assert.Equal(t, mockUser, user)
}

func TestGetUser_UserNotFound(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	username := "nonexistentuser"

	user, err := service.GetUser(username)

	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Empty(t, user.Username)
}

func TestDeleteUser_Success(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	username := "testuser"
	mockUser := models.User{
		Username: username,
	}

	store.UpsertUser(mockUser)

	deleted, err := service.DeleteUser(username)

	assert.NoError(t, err)
	assert.True(t, deleted)
}

func TestDeleteUser_UserNotFound(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	username := "nonexistentuser"

	deleted, err := service.DeleteUser(username)

	assert.NoError(t, err)
	assert.False(t, deleted)
}

func TestChangePassword_Success(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	username := "testuser"
	oldPassword := "oldpassword"
	newPassword := "newpassword"

	hashedPassword, _ := hashPassword(oldPassword)
	mockUser := models.User{
		Username:       username,
		HashedPassword: hashedPassword,
	}
	store.UpsertUser(mockUser)

	success, err := service.ChangePassword(username, oldPassword, newPassword)

	assert.NoError(t, err)
	assert.True(t, success)

	updatedUser, _ := store.UserByUsername(username)
	assert.True(t, checkPasswordHash(newPassword, updatedUser.HashedPassword))
}

func TestChangePassword_UserNotFound(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	username := "nonexistentuser"
	oldPassword := "oldpassword"
	newPassword := "newpassword"

	success, err := service.ChangePassword(username, oldPassword, newPassword)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.False(t, success)
}

func TestChangePassword_InvalidOldPassword(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	username := "testuser"
	oldPassword := "oldpassword"
	newPassword := "newpassword"
	wrongOldPassword := "wrongpassword"

	hashedPassword, _ := hashPassword(oldPassword)
	mockUser := models.User{
		Username:       username,
		HashedPassword: hashedPassword,
	}
	store.UpsertUser(mockUser)

	success, err := service.ChangePassword(username, wrongOldPassword, newPassword)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPassword)
	assert.False(t, success)
}

func TestRefreshToken_Success(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	credentials := models.Credentials{
		Username: "testuser",
		Password: "password",
	}

	hashedPassword, _ := hashPassword(credentials.Password)
	mockUser := models.User{
		Username:       credentials.Username,
		HashedPassword: hashedPassword,
	}
	store.UpsertUser(mockUser)

	// Login to get initial token
	token, err := service.Login(credentials)
	assert.NoError(t, err)

	// Refresh token
	newToken, err := service.RefreshToken(token)
	assert.NoError(t, err)
	assert.NotEqual(t, token.Value, newToken.Value)
	assert.NotEqual(t, token.RefreshToken, newToken.RefreshToken)
	assert.NotZero(t, newToken.Value)
	assert.NotZero(t, newToken.RefreshToken)

	// Verify old token is invalid
	user, err := service.GetUsernameByToken(token)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Zero(t, user)

	// Verify new token works
	user, err = service.GetUsernameByToken(newToken)
	assert.NoError(t, err)
	assert.Equal(t, mockUser.Username, user)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	// Create expired token
	expiredToken := models.Token{
		Value:          "expiredtoken",
		RefreshToken:   "expiredrefreshtoken",
		Expires:        time.Now().Add(-time.Hour),
		RefreshExpires: time.Now().Add(-time.Hour),
	}

	// Try to refresh
	newToken, err := service.RefreshToken(expiredToken)
	assert.Error(t, err)
	assert.Zero(t, newToken)
}

func TestRefreshToken_RevokedToken(t *testing.T) {
	store, err := storage.NewUsersStorage(testutil.NewMemoryStorage())
	assert.NoError(t, err)
	service := NewService(store)

	credentials := models.Credentials{
		Username: "testuser",
		Password: "password",
	}

	hashedPassword, _ := hashPassword(credentials.Password)
	mockUser := models.User{
		Username:       credentials.Username,
		HashedPassword: hashedPassword,
	}
	store.UpsertUser(mockUser)

	// Login to get initial token
	token, err := service.Login(credentials)
	assert.NoError(t, err)

	// Revoke the refresh token
	err = store.RevokeRefreshToken(token.RefreshToken)
	assert.NoError(t, err)

	// Try to refresh
	newToken, err := service.RefreshToken(token)
	assert.Error(t, err)
	assert.Zero(t, newToken)
}
