package storage

import (
	"testing"
	"time"

	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var user = models.User{
	Username:       "testuser",
	HashedPassword: "hashedpassword",
}

func setupUserStorage(t *testing.T) (UserStorage, *gorm.DB) {
	db, err := NewGormDb(":memory:", 0o000)
	if err != nil {
		t.Fatalf("new db: %v", err)
	}
	userStore, err := NewUsersStorage(db)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	return userStore, db
}

func TestUserStorage_Migrates(t *testing.T) {
	_, db := setupUserStorage(t)
	// ensure migrations created the deployments table
	has := db.Migrator().HasTable(&models.User{})
	assert.True(t, has)
}

func TestHasUsers(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Test when there are no users
	exists, err := s.HasUsers()
	assert.NoError(t, err)
	assert.False(t, exists)

	// Add a user to the database
	_, err = s.UpsertUser(user)
	assert.NoError(t, err)

	// Test when there are users
	exists, err = s.HasUsers()
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUserByUsername(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Test when the username does not exist
	user, err := s.UserByUsername("nonexistentuser")
	assert.NoError(t, err)
	assert.Equal(t, models.User{}, user)

	// Add a user to the database
	userToAdd := models.User{
		Username:       "testuser",
		HashedPassword: "hashedpassword",
	}
	_, err = s.UpsertUser(userToAdd)
	assert.NoError(t, err)

	// Test when the username exists
	user, err = s.UserByUsername("testuser")
	assert.NoError(t, err)
	assert.Equal(t, userToAdd.Username, user.Username)
	assert.Equal(t, userToAdd.HashedPassword, user.HashedPassword)
}

func TestUpsertUser_InsertNewUser(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Create a new user
	user := models.User{
		Username:       "newuser",
		HashedPassword: "newpassword",
	}

	// Insert the new user
	insertedUser, err := s.UpsertUser(user)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, insertedUser.Username)
	assert.Equal(t, user.HashedPassword, insertedUser.HashedPassword)

	// Verify the user was inserted
	fetchedUser, err := s.UserByUsername("newuser")
	assert.NoError(t, err)
	assert.Equal(t, user.Username, fetchedUser.Username)
	assert.Equal(t, user.HashedPassword, fetchedUser.HashedPassword)
}

func TestUpsertUser_UpdateExistingUser(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Insert an initial user
	initialUser := models.User{
		Username:       "existinguser",
		HashedPassword: "initialpassword",
	}
	_, err := s.UpsertUser(initialUser)
	assert.NoError(t, err)

	// Update the user
	updatedUser := models.User{
		Username:       "existinguser",
		HashedPassword: "updatedpassword",
	}
	upsertedUser, err := s.UpsertUser(updatedUser)
	assert.NoError(t, err)
	assert.Equal(t, updatedUser.Username, upsertedUser.Username)
	assert.Equal(t, updatedUser.HashedPassword, upsertedUser.HashedPassword)

	// Verify the user was updated
	fetchedUser, err := s.UserByUsername("existinguser")
	assert.NoError(t, err)
	assert.Equal(t, updatedUser.Username, fetchedUser.Username)
	assert.Equal(t, updatedUser.HashedPassword, fetchedUser.HashedPassword)
}

func TestUpsertUser_InvalidUser(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Attempt to insert a user with invalid data
	invalidUser := models.User{
		Username:       "",
		HashedPassword: "",
	}
	_, err := s.UpsertUser(invalidUser)
	assert.ErrorIs(t, err, ErrEmptyUsername)
}

func TestDeleteUserByUserName_Success(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Add a user to the database
	userToAdd := models.User{
		Username:       "testuser",
		HashedPassword: "hashedpassword",
	}
	_, err := s.UpsertUser(userToAdd)
	assert.NoError(t, err)

	// Delete the user by username
	deleted, err := s.DeleteUserByUserName("testuser")
	assert.NoError(t, err)
	assert.True(t, deleted)

	// Verify the user was deleted
	user, err := s.UserByUsername("testuser")
	assert.NoError(t, err)
	assert.Equal(t, models.User{}, user)
}

func TestDeleteUserByUserName_NotFound(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Attempt to delete a non-existent user
	deleted, err := s.DeleteUserByUserName("nonexistentuser")
	assert.NoError(t, err)
	assert.False(t, deleted)
}

func TestNewSession(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Create a test token
	token := models.Token{
		RefreshToken:   "test_refresh_token",
		RefreshExpires: time.Now().Add(time.Hour),
	}

	// Create a new session
	session, err := s.NewSession(token, "testuser")
	assert.NoError(t, err)
	assert.Equal(t, "test_refresh_token", session.RefreshToken)
	assert.Equal(t, token.RefreshExpires, session.RefreshExpires)
	assert.Equal(t, "testuser", session.Username)
	assert.False(t, session.Revoked)

	// Verify the session was created
	storedSession, err := s.SessionByRefreshToken("test_refresh_token")
	assert.NoError(t, err)
	assert.Equal(t, session.SessionID, storedSession.SessionID)
}

func TestSessionByRefreshToken(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Test non-existent token
	_, err := s.SessionByRefreshToken("nonexistent_token")
	assert.ErrorIs(t, err, ErrNotFound)

	// Create a test session
	token := models.Token{
		RefreshToken:   "test_refresh_token",
		RefreshExpires: time.Now().Add(time.Hour),
	}
	_, err = s.NewSession(token, "testuser")
	assert.NoError(t, err)

	// Test existing token
	session, err := s.SessionByRefreshToken("test_refresh_token")
	assert.NoError(t, err)
	assert.Equal(t, "test_refresh_token", session.RefreshToken)
	assert.Equal(t, "testuser", session.Username)
}

func TestRevokeRefreshToken(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Test non-existent token
	err := s.RevokeRefreshToken("nonexistent_token")
	assert.ErrorIs(t, err, ErrNotFound)

	// Create a test session
	token := models.Token{
		RefreshToken:   "test_refresh_token",
		RefreshExpires: time.Now().Add(time.Hour),
	}
	_, err = s.NewSession(token, "testuser")
	assert.NoError(t, err)

	// Revoke the token
	err = s.RevokeRefreshToken("test_refresh_token")
	assert.NoError(t, err)

	// Verify the token was revoked
	session, err := s.SessionByRefreshToken("test_refresh_token")
	assert.NoError(t, err)
	assert.True(t, session.Revoked)
}

func TestRevokeAllUserSessions(t *testing.T) {
	s, _ := setupUserStorage(t)

	// Create test sessions for the same user
	token1 := models.Token{
		RefreshToken:   "token1",
		RefreshExpires: time.Now().Add(time.Hour),
	}
	token2 := models.Token{
		RefreshToken:   "token2",
		RefreshExpires: time.Now().Add(time.Hour),
	}
	_, err := s.NewSession(token1, "testuser")
	assert.NoError(t, err)
	_, err = s.NewSession(token2, "testuser")
	assert.NoError(t, err)

	// Revoke all sessions for the user
	err = s.RevokeAllUserSessions("testuser")
	assert.NoError(t, err)

	// Verify both sessions were revoked
	session1, err := s.SessionByRefreshToken("token1")
	assert.NoError(t, err)
	assert.True(t, session1.Revoked)

	session2, err := s.SessionByRefreshToken("token2")
	assert.NoError(t, err)
	assert.True(t, session2.Revoked)
}
