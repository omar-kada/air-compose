package mappers

import (
	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/models"
)

// SettingsMapper maps models.Settings to api.Settings
type SettingsMapper struct{}

// Map converts a models.Settings to an api.Settings
func (SettingsMapper) Map(settings models.Settings) api.Settings {
	token := settings.GetObfuscatedToken()
	notificationURL := settings.GetObfuscatedNotificationURL()
	return api.Settings{
		Repo:              settings.Git.Repo,
		Branch:            &settings.Git.Branch,
		Cron:              &settings.Schedule.Cron,
		Token:             &token,
		Username:          &settings.Git.Username,
		NotificationURL:   &notificationURL,
		NotificationTypes: mapEventTypes(settings.Notifications.NotificationTypes),
		Oidc:              mapOidcConfig(settings.Oidc),
	}
}

func mapEventTypes(types []models.EventType) []api.EventType {
	if types == nil {
		return nil
	}
	eventTypes := make([]api.EventType, len(types))
	for i, et := range types {
		eventTypes[i] = api.EventType(et)
	}
	return eventTypes
}

func mapOidcConfig(config models.OidcConfig) *api.OidcSettings {
	res := api.OidcSettings{
		IssuerURL:    config.IssuerURL,
		ClientID:     config.ClientID,
		ClientSecret: config.GetObfuscatedClientSecret(),
	}
	return &res
}

// UnMap transforms back from api.Settings to models.Settings
func (SettingsMapper) UnMap(settings api.Settings) models.Settings {
	res := models.Settings{
		Git: models.GitConfig{
			Repo: settings.Repo,
		},
		Notifications: models.NotificationConfig{
			NotificationTypes: unmapEventTypes(settings.NotificationTypes),
		},
	}
	if settings.Branch != nil {
		res.Git.Branch = *settings.Branch
	}
	if settings.Cron != nil {
		res.Schedule.Cron = *settings.Cron
	}
	if settings.Token != nil {
		res.Git.Token = *settings.Token
	}
	if settings.Username != nil {
		res.Git.Username = *settings.Username
	}
	if settings.NotificationURL != nil {
		res.Notifications.NotificationURL = *settings.NotificationURL
	}
	if settings.Oidc != nil {
		res.Oidc = unmapOidcConfig(*settings.Oidc)
	}
	return res
}

func unmapEventTypes(types []api.EventType) []models.EventType {
	if types == nil {
		return nil
	}
	eventTypes := make([]models.EventType, len(types))
	for i, et := range types {
		eventTypes[i] = models.EventType(et)
	}
	return eventTypes
}

func unmapOidcConfig(config api.OidcSettings) models.OidcConfig {
	return models.OidcConfig{
		IssuerURL:    config.IssuerURL,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
	}
}
