package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/process"
	"omar-kada/air-compose/internal/server/mappers"
	"omar-kada/air-compose/internal/server/middlewares"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/internal/users"
	"omar-kada/air-compose/models"
)

var (
	errUserNotFound  = errors.New("user error")
	errShouldntReach = errors.New("shouldn't be reachable")
)

// Handler implements the generated strict server interface
type Handler struct {
	configStore    storage.ConfigStore
	processService process.Service
	accountService users.AccountService

	depMapper        mappers.DeploymentMapper
	depDetailsMapper mappers.DeploymentDetailsMapper
	eventMapper      mappers.EventMapper
	diffMapper       mappers.DiffMapper
	statusMapper     mappers.StatusMapper
	stateMapper      mappers.StateMapper
	configMapper     mappers.ConfigMapper
	settingsMapper   mappers.SettingsMapper
	featuresMapper   mappers.FeaturesMapper
}

// NewHandler creates a new Handler
func NewHandler(configStore storage.ConfigStore, processService process.Service, userService users.AccountService) *Handler {
	diffMapper := mappers.DiffMapper{}
	eventMapper := mappers.EventMapper{}

	return &Handler{
		configStore:      configStore,
		processService:   processService,
		accountService:   userService,
		depMapper:        mappers.NewDeploymentMapper(),
		depDetailsMapper: mappers.NewDeploymentDetailsMapper(diffMapper, eventMapper),
		eventMapper:      eventMapper,
		diffMapper:       diffMapper,
		statusMapper:     mappers.StatusMapper{},
		stateMapper:      mappers.StateMapper{},
		configMapper:     mappers.ConfigMapper{},
	}
}

// DeployementAPIList lists deployments with pagination support
func (h *Handler) DeployementAPIList(_ context.Context, request api.DeployementAPIListRequestObject) (api.DeployementAPIListResponseObject, error) {
	offset, err := validateCursorOffset(request.Params.Offset)
	if err != nil {
		return nil, fmt.Errorf("invalid after value")
	}

	if request.Params.Limit <= 0 {
		return nil, fmt.Errorf("invalid first value")
	}

	deps, err := h.processService.GetDeployments(int(request.Params.Limit), offset)

	return api.DeployementAPIList200JSONResponse{
		Items:    models.ListMapper(h.depMapper.Map)(deps),
		PageInfo: h.depMapper.MapToPageInfo(deps, int(request.Params.Limit)),
	}, err
}

func validateCursorOffset(offsetStr *string) (uint64, error) {
	offset := uint64(0)
	var err error
	if offsetStr != nil && *offsetStr != "" {
		offset, err = strconv.ParseUint(*offsetStr, 10, 64)
	}
	return offset, err
}

// DeployementAPIRead retrieves details of a specific deployment
func (h *Handler) DeployementAPIRead(_ context.Context, request api.DeployementAPIReadRequestObject) (api.DeployementAPIReadResponseObject, error) {
	id, err := strconv.ParseUint(request.Id, 10, 64)
	if err != nil {
		return nil, err
	}
	dep, err := h.processService.GetDeployment(id)
	if err != nil {
		return nil, err
	} else if dep.ID == 0 {
		return api.DeployementAPIReaddefaultJSONResponse{
			Body: api.Error{
				Code:    api.ErrorCodeNOTFOUND,
				Message: err.Error(),
			},
			StatusCode: http.StatusNotFound,
		}, err
	}

	return api.DeployementAPIRead200JSONResponse(h.depDetailsMapper.Map(dep)), err
}

// DeployementAPISync syncs the deployment
func (h *Handler) DeployementAPISync(_ context.Context, _ api.DeployementAPISyncRequestObject) (api.DeployementAPISyncResponseObject, error) {
	dep, err := h.processService.SyncDeployment()
	if err != nil {
		slog.Error(err.Error())
	} else if reflect.DeepEqual(models.Deployment{}, dep) {
		return api.DeployementAPISync204Response{}, nil
	}
	return api.DeployementAPISync200JSONResponse(h.depDetailsMapper.Map(dep)), err
}

// StatusAPIGet retrieves the status of managed stacks
func (h *Handler) StatusAPIGet(_ context.Context, _ api.StatusAPIGetRequestObject) (api.StatusAPIGetResponseObject, error) {
	stacks, err := h.processService.GetManagedStacks()
	if err != nil {
		return nil, err
	}

	result := models.MapMapper[string](
		models.ListMapper(h.statusMapper.Map),
	)(stacks)

	var response []api.StackStatus
	for stackName, containers := range result {
		response = append(response, api.StackStatus{
			StackId:  stackName,
			Name:     stackName,
			Services: containers,
		})
	}
	return api.StatusAPIGet200JSONResponse(response), nil
}

// StateAPIGet retrieves the state of AirCompose
func (h *Handler) StateAPIGet(_ context.Context, _ api.StateAPIGetRequestObject) (api.StateAPIGetResponseObject, error) {
	state, err := h.processService.GetCurrentState()
	if err != nil {
		return nil, err
	}
	return api.StateAPIGet200JSONResponse(h.stateMapper.Map(state)), nil
}

// DiffAPIGet retrieves the differences in files
func (h *Handler) DiffAPIGet(_ context.Context, _ api.DiffAPIGetRequestObject) (api.DiffAPIGetResponseObject, error) {
	fileDiffs, err := h.processService.GetDiff()
	if err != nil {
		return nil, err
	}
	return api.DiffAPIGet200JSONResponse(models.ListMapper(h.diffMapper.Map)(fileDiffs)), nil
}

// ConfigAPIGet retrieves the current configuration
func (h *Handler) ConfigAPIGet(_ context.Context, _ api.ConfigAPIGetRequestObject) (api.ConfigAPIGetResponseObject, error) {
	config, err := h.configStore.Get()
	if err != nil {
		return nil, err
	}
	return api.ConfigAPIGet200JSONResponse(h.configMapper.Map(config)), nil
}

// ConfigAPISet updates the current configuration
func (h *Handler) ConfigAPISet(_ context.Context, r api.ConfigAPISetRequestObject) (api.ConfigAPISetResponseObject, error) {
	config := h.configMapper.UnMap(api.Config(*r.Body))
	oldConfig, err := h.configStore.Get()
	if err != nil {
		return nil, err
	}
	oldConfig.Environment = config.Environment
	oldConfig.Services = config.Services
	err = h.configStore.Update(oldConfig)
	if err != nil {
		return nil, err
	}
	return api.ConfigAPISet200JSONResponse(h.configMapper.Map(oldConfig)), nil
}

// SettingsAPIGet retrieves the current settings
func (h *Handler) SettingsAPIGet(_ context.Context, _ api.SettingsAPIGetRequestObject) (api.SettingsAPIGetResponseObject, error) {
	config, err := h.configStore.Get()
	if err != nil {
		return nil, err
	}
	return api.SettingsAPIGet200JSONResponse(h.settingsMapper.Map(config.Settings)), nil
}

// SettingsAPISet updates the current settings
func (h *Handler) SettingsAPISet(_ context.Context, r api.SettingsAPISetRequestObject) (api.SettingsAPISetResponseObject, error) {
	oldConfig, err := h.configStore.Get()
	if err != nil {
		return nil, err
	}
	settings := h.settingsMapper.UnMap(api.Settings(*r.Body))
	oldConfig.Settings = settings
	err = h.configStore.Update(oldConfig)
	if err != nil {
		return nil, err
	}
	return api.SettingsAPISet200JSONResponse(h.settingsMapper.Map(settings)), nil
}

// SettingsAPITestGitConnection tests the connection to a Git repository
func (h *Handler) SettingsAPITestGitConnection(_ context.Context, r api.SettingsAPITestGitConnectionRequestObject) (api.SettingsAPITestGitConnectionResponseObject, error) {
	res, err := h.processService.TestGitConnection(r.Body.Repo, *r.Body.Branch, *r.Body.Username, *r.Body.Token)
	return api.SettingsAPITestGitConnection200JSONResponse{
		Success: res,
	}, err
}

// FeaturesAPIGet retrieves the current features
func (h *Handler) FeaturesAPIGet(_ context.Context, _ api.FeaturesAPIGetRequestObject) (api.FeaturesAPIGetResponseObject, error) {
	return api.FeaturesAPIGet200JSONResponse(h.featuresMapper.Map(models.LoadFeatures())), nil
}

// AuthAPIRegister registers a new user
func (*Handler) AuthAPIRegister(_ context.Context, _ api.AuthAPIRegisterRequestObject) (api.AuthAPIRegisterResponseObject, error) {
	// should be done in the auth middleware so if we react this return an error
	return api.AuthAPIRegister200JSONResponse{}, errShouldntReach
}

// AuthAPILogin logs in a user
func (*Handler) AuthAPILogin(_ context.Context, _ api.AuthAPILoginRequestObject) (api.AuthAPILoginResponseObject, error) {
	// should be done in the auth middleware so if we react this return an error
	return api.AuthAPILogin200JSONResponse{}, errShouldntReach
}

// AuthAPIRefresh refreshes token
func (*Handler) AuthAPIRefresh(_ context.Context, _ api.AuthAPIRefreshRequestObject) (api.AuthAPIRefreshResponseObject, error) {
	// should be done in the auth middleware so if we react this return an error
	return api.AuthAPIRefresh200JSONResponse{}, errShouldntReach
}

// AuthAPILogout logs out a user
func (*Handler) AuthAPILogout(_ context.Context, _ api.AuthAPILogoutRequestObject) (api.AuthAPILogoutResponseObject, error) {
	// should be done in the auth middleware so if we react this return an error
	return api.AuthAPILogout200JSONResponse{}, errShouldntReach
}

// AuthAPIRegistered checks if a user is registered
func (*Handler) AuthAPIRegistered(_ context.Context, _ api.AuthAPIRegisteredRequestObject) (api.AuthAPIRegisteredResponseObject, error) {
	// should be done in the auth middleware so if we react this return an error
	return api.AuthAPIRegistereddefaultJSONResponse{}, errShouldntReach
}

// UserAPIGet returns the authenticated user's information
func (*Handler) UserAPIGet(ctx context.Context, _ api.UserAPIGetRequestObject) (api.UserAPIGetResponseObject, error) {
	username, exists := middlewares.UsernameFromContext(ctx)
	if !exists {
		return nil, nil
	}

	return api.UserAPIGet200JSONResponse{
		Username: username,
	}, nil
}

// UserAPIDelete deletes the authenticated user
func (h *Handler) UserAPIDelete(ctx context.Context, _ api.UserAPIDeleteRequestObject) (api.UserAPIDeleteResponseObject, error) {
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
func (h *Handler) UserAPIChangePassword(ctx context.Context, r api.UserAPIChangePasswordRequestObject) (api.UserAPIChangePasswordResponseObject, error) {
	username, exists := middlewares.UsernameFromContext(ctx)
	if !exists {
		return api.UserAPIChangePassworddefaultJSONResponse{}, errUserNotFound
	}
	ok, err := h.accountService.ChangePassword(username, r.Body.OldPass, r.Body.NewPass)
	return api.UserAPIChangePassword200JSONResponse{
		Success: ok,
	}, err
}

// NotificationsAPIList lists notifications with pagination support
func (h *Handler) NotificationsAPIList(_ context.Context, request api.NotificationsAPIListRequestObject) (api.NotificationsAPIListResponseObject, error) {
	offset, err := validateCursorOffset(request.Params.Offset)
	if err != nil {
		return nil, fmt.Errorf("invalid after value")
	}

	if request.Params.Limit <= 0 {
		return nil, fmt.Errorf("invalid first value")
	}

	events, err := h.processService.GetNotifications(int(request.Params.Limit), offset)

	return api.NotificationsAPIList200JSONResponse{
		Items:    models.ListMapper(h.eventMapper.Map)(events),
		PageInfo: h.eventMapper.MapToPageInfo(events, int(request.Params.Limit)),
	}, err
}
