package mappers

import (
	"testing"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
)

func TestUserMapper_Map(t *testing.T) {
	// Setup
	userMapper := UserMapper{}

	// Test data
	user := models.User{
		Username: "testUser",
		Type:     models.UserTypeLocal,
	}

	// Expected result
	expected := api.User{
		Username: user.Username,
		Type:     api.UserTypeLOCAL,
	}

	// Execute
	actual := userMapper.Map(user)

	// Assert
	assert.Equal(t, expected, actual)
}
