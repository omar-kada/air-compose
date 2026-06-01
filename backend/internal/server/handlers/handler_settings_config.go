package handlers

import (
	"context"
	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/server/mappers"
	"omar-kada/air-compose/internal/storage"
)

// SettingsConfigHandler handles settings and configuration-related operations.
type SettingsConfigHandler struct {
	configStore    storage.ConfigStore
	configMapper   mappers.MapperUnmapper[models.Config, api.Config]
	settingsMapper mappers.MapperUnmapper[models.Settings, api.Settings]
	featuresMapper mappers.Mapper[models.Features, api.Features]
	fetcher        git.Fetcher
}

// FeaturesAPIGet gets the features.
func (h *SettingsConfigHandler) FeaturesAPIGet(_ context.Context, _ api.FeaturesAPIGetRequestObject) (api.FeaturesAPIGetResponseObject, error) {
	return api.FeaturesAPIGet200JSONResponse(h.featuresMapper.Map(models.LoadFeatures())), nil
}

// ConfigAPIGet retrieves the current configuration
func (h *SettingsConfigHandler) ConfigAPIGet(_ context.Context, _ api.ConfigAPIGetRequestObject) (api.ConfigAPIGetResponseObject, error) {
	config := h.configStore.Get()
	return api.ConfigAPIGet200JSONResponse(h.configMapper.Map(config)), nil
}

// ConfigAPISet updates the current configuration
func (h *SettingsConfigHandler) ConfigAPISet(_ context.Context, r api.ConfigAPISetRequestObject) (api.ConfigAPISetResponseObject, error) {
	config := h.configMapper.UnMap(api.Config(*r.Body))
	oldConfig := h.configStore.Get()
	oldConfig.Environment = config.Environment
	oldConfig.Services = config.Services
	err := h.configStore.Update(oldConfig)
	if err != nil {
		return nil, err
	}
	return api.ConfigAPISet200JSONResponse(h.configMapper.Map(oldConfig)), nil
}

// SettingsAPIGet retrieves the current settings
func (h *SettingsConfigHandler) SettingsAPIGet(_ context.Context, _ api.SettingsAPIGetRequestObject) (api.SettingsAPIGetResponseObject, error) {
	config := h.configStore.Get()
	return api.SettingsAPIGet200JSONResponse(h.settingsMapper.Map(config.Settings)), nil
}

// SettingsAPISet updates the current settings
func (h *SettingsConfigHandler) SettingsAPISet(_ context.Context, r api.SettingsAPISetRequestObject) (api.SettingsAPISetResponseObject, error) {
	oldConfig := h.configStore.Get()
	settings := h.settingsMapper.UnMap(api.Settings(*r.Body))
	oldConfig.Settings = settings
	err := h.configStore.Update(oldConfig)
	if err != nil {
		return nil, err
	}
	return api.SettingsAPISet200JSONResponse(h.settingsMapper.Map(settings)), nil
}

// SettingsAPITestGitConnection tests the connection to a Git repository
func (h *SettingsConfigHandler) SettingsAPITestGitConnection(_ context.Context, r api.SettingsAPITestGitConnectionRequestObject) (api.SettingsAPITestGitConnectionResponseObject, error) {
	res, err := h.fetcher.TestGitConnection(r.Body.Repo, *r.Body.Branch, *r.Body.Username, *r.Body.Token)
	return api.SettingsAPITestGitConnection200JSONResponse{
		Success: res,
	}, err
}
