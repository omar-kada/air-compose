package mappers

import (
	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"
)

// UserMapper maps between domain and API user models.
type UserMapper struct{}

// Map converts a domain User model to an API User model.
func (UserMapper) Map(user models.User) api.User {
	return api.User{
		Username: user.Username,
		Type:     api.UserType(user.Type),
	}
}
