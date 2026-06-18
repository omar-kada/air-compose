package mappers

import (
	"testing"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestSettingsMapper_Map(t *testing.T) {
	main := "main"
	cron := "0 0 * * *"
	username := "user"
	token := "123456789123456789"
	notificationURL := "gotify://123456789"
	obfuscatedToken := models.Obfuscate(token)
	obfuscatedURL := models.Obfuscate(notificationURL)
	empty := ""
	oidc := models.OidcConfig{
		IssuerURL:    "issuer",
		ClientID:     "client",
		ClientSecret: "secret",
	}
	oidcAPI := api.OidcSettings{
		IssuerURL:    "issuer",
		ClientID:     "client",
		ClientSecret: "******************************",
	}
	cases := []struct {
		name string
		in   models.Settings
		want api.Settings
	}{
		{
			name: "basic",
			in: models.Settings{
				Git: models.GitConfig{
					Repo:     "https://github.com/example/repo",
					Branch:   main,
					Username: username,
					Token:    token,
				},
				Schedule: models.ScheduleConfig{
					Cron:               cron,
					RetriesOnUnhealthy: 3,
					RetryDelay:         30000,
				},
				Notifications: models.NotificationConfig{
					NotificationURL:   notificationURL,
					NotificationTypes: []models.EventType{},
				},
				Oidc: oidc,
			},
			want: api.Settings{
				Repo:               "https://github.com/example/repo",
				Branch:             &main,
				Cron:               &cron,
				Token:              &obfuscatedToken,
				Username:           &username,
				NotificationURL:    &obfuscatedURL,
				NotificationTypes:  []api.EventType{},
				Oidc:               &oidcAPI,
				RetriesOnUnhealthy: 3,
				RetryDelay:         30000,
			},
		},
		{
			name: "empty",
			in: models.Settings{
				Git: models.GitConfig{
					Repo: "",
				},
				Notifications: models.NotificationConfig{
					NotificationTypes: []models.EventType{},
				},
				Oidc: models.OidcConfig{},
			},
			want: api.Settings{
				Repo:               "",
				Branch:             &empty,
				Cron:               &empty,
				Token:              &empty,
				Username:           &empty,
				NotificationURL:    &empty,
				NotificationTypes:  []api.EventType{},
				Oidc:               &api.OidcSettings{},
				RetriesOnUnhealthy: 0,
				RetryDelay:         0,
			},
		},
	}

	m := SettingsMapper{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.Map(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSettingsMapper_UnMap(t *testing.T) {
	branch := "main"
	cron := "0 0 * * *"
	repo := "https://github.com/example/repo"
	username := "user"
	token := "123456789123456789"
	notificationURL := "gotify://123456789"

	oidc := api.OidcSettings{
		IssuerURL:    "issuer",
		ClientID:     "client",
		ClientSecret: "secret",
	}
	oidcModel := models.OidcConfig{
		IssuerURL:    "issuer",
		ClientID:     "client",
		ClientSecret: "secret",
	}
	cases := []struct {
		name string
		in   api.Settings
		want models.Settings
	}{
		{
			name: "basic",
			in: api.Settings{
				Repo:               repo,
				Branch:             &branch,
				Cron:               &cron,
				Username:           &username,
				Token:              &token,
				NotificationURL:    &notificationURL,
				NotificationTypes:  []api.EventType{},
				Oidc:               &oidc,
				RetriesOnUnhealthy: 3,
				RetryDelay:         30000,
			},
			want: models.Settings{
				Git: models.GitConfig{
					Repo:     repo,
					Branch:   branch,
					Username: username,
					Token:    token,
				},
				Schedule: models.ScheduleConfig{
					Cron:               cron,
					RetriesOnUnhealthy: 3,
					RetryDelay:         30000,
				},
				Notifications: models.NotificationConfig{
					NotificationURL:   notificationURL,
					NotificationTypes: []models.EventType{},
				},
				Oidc: oidcModel,
			},
		},
		{
			name: "empty",
			in: api.Settings{
				Branch:             nil,
				Cron:               nil,
				Repo:               "",
				Username:           nil,
				Token:              nil,
				NotificationURL:    nil,
				NotificationTypes:  []api.EventType{},
				Oidc:               &api.OidcSettings{},
				RetriesOnUnhealthy: 0,
				RetryDelay:         0,
			},
			want: models.Settings{
				Git: models.GitConfig{
					Repo:     "",
					Branch:   "",
					Username: "",
					Token:    "",
				},
				Schedule: models.ScheduleConfig{
					Cron:               "",
					RetriesOnUnhealthy: 0,
				},
				Notifications: models.NotificationConfig{
					NotificationURL:   "",
					NotificationTypes: []models.EventType{},
				},
				Oidc: models.OidcConfig{},
			},
		},
	}

	m := SettingsMapper{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.UnMap(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
