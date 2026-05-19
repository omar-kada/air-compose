package handlers

import (
	"context"
	"errors"
	"fmt"
	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/server/mappers"
	"omar-kada/air-compose/internal/server/middlewares"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/internal/users"
)

var (
	errUserNotFound  = errors.New("user error")
	errShouldntReach = errors.New("shouldn't be reachable")
)

// AuthUserHandler handles authentication and user-related operations.
type AuthUserHandler struct {
	accountService users.AccountService
	configStore    storage.ConfigStore
	userMapper     mappers.UserMapper
}

// AuthAPIRegister registers a new user
func (*AuthUserHandler) AuthAPIRegister(_ context.Context, _ api.AuthAPIRegisterRequestObject) (api.AuthAPIRegisterResponseObject, error) {
	// should be done in the auth middleware so if we reach this return an error
	return api.AuthAPIRegister200JSONResponse{}, errShouldntReach
}

// AuthAPILogin logs in a user
func (*AuthUserHandler) AuthAPILogin(_ context.Context, _ api.AuthAPILoginRequestObject) (api.AuthAPILoginResponseObject, error) {
	// should be done in the auth middleware so if we reach this return an error
	return api.AuthAPILogin200JSONResponse{}, errShouldntReach
}

// AuthAPIRefresh refreshes token
func (*AuthUserHandler) AuthAPIRefresh(_ context.Context, _ api.AuthAPIRefreshRequestObject) (api.AuthAPIRefreshResponseObject, error) {
	// should be done in the auth middleware so if we reach this return an error
	return api.AuthAPIRefresh200JSONResponse{}, errShouldntReach
}

// AuthAPILogout logs out a user
func (*AuthUserHandler) AuthAPILogout(_ context.Context, _ api.AuthAPILogoutRequestObject) (api.AuthAPILogoutResponseObject, error) {
	// should be done in the auth middleware so if we reach this return an error
	return api.AuthAPILogout200JSONResponse{}, errShouldntReach
}

// AuthAPIRegistered checks if a user is registered
func (h *AuthUserHandler) AuthAPIRegistered(_ context.Context, _ api.AuthAPIRegisteredRequestObject) (api.AuthAPIRegisteredResponseObject, error) {
	hasUsers, err := h.accountService.IsRegistered()
	if err != nil {
		return api.AuthAPIRegistereddefaultJSONResponse{}, err
	}
	cfg, err := h.configStore.Get()
	if err != nil {
		return api.AuthAPIRegistereddefaultJSONResponse{}, err
	}
	return api.AuthAPIRegistered200JSONResponse{
		Registered: hasUsers,
		Oidc:       cfg.Settings.Oidc.IssuerURL != "",
	}, nil
}

// OIDCAPIOidcCallback handles the OIDC callback after authentication
func (*AuthUserHandler) OIDCAPIOidcCallback(_ context.Context, _ api.OIDCAPIOidcCallbackRequestObject) (api.OIDCAPIOidcCallbackResponseObject, error) {
	// should be done in the oidc middleware so if we reach this return an error
	return api.OIDCAPIOidcCallbackdefaultJSONResponse{}, errShouldntReach
}

// OIDCAPIOidcLogin initiates the OIDC login flow
func (*AuthUserHandler) OIDCAPIOidcLogin(_ context.Context, _ api.OIDCAPIOidcLoginRequestObject) (api.OIDCAPIOidcLoginResponseObject, error) {
	// should be done in the oidc middleware so if we reach this return an error
	return api.OIDCAPIOidcLogindefaultJSONResponse{}, errShouldntReach
}

// UserAPIGet returns the authenticated user's information
func (h *AuthUserHandler) UserAPIGet(ctx context.Context, _ api.UserAPIGetRequestObject) (api.UserAPIGetResponseObject, error) {
	username, exists := middlewares.UsernameFromContext(ctx)
	if !exists || username == "" {
		return nil, nil
	}

	user, err := h.accountService.GetUser(username)

	return api.UserAPIGet200JSONResponse(h.userMapper.Map(user)), err
}

// UserAPIDelete deletes the authenticated user
func (h *AuthUserHandler) UserAPIDelete(ctx context.Context, _ api.UserAPIDeleteRequestObject) (api.UserAPIDeleteResponseObject, error) {
	username, exists := middlewares.UsernameFromContext(ctx)
	if !exists {
		return api.UserAPIDeletedefaultJSONResponse{}, errUserNotFound
	}
	ok, err := h.accountService.DeleteUser(username)
	if err != nil || !ok {
		return api.UserAPIDelete200JSONResponse{
			Success: false,
		}, fmt.Errorf("failed to delete user: %w", err)
	}
	return api.UserAPIDelete200JSONResponse{
		Success: true,
	}, nil
}

// UserAPIChangePassword changes the password for the authenticated user
func (h *AuthUserHandler) UserAPIChangePassword(ctx context.Context, r api.UserAPIChangePasswordRequestObject) (api.UserAPIChangePasswordResponseObject, error) {
	username, exists := middlewares.UsernameFromContext(ctx)
	if !exists {
		return api.UserAPIChangePassworddefaultJSONResponse{}, errUserNotFound
	}
	ok, err := h.accountService.ChangePassword(username, r.Body.OldPass, r.Body.NewPass)
	return api.UserAPIChangePassword200JSONResponse{
		Success: ok,
	}, err
}
