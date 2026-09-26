package models

import (
	"reflect"
	"strings"
	"testing"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/stretchr/testify/assert"
)

func TestConfigPerService_BuildsCorrectArray(t *testing.T) {
	cfg := Config{
		Environment: Environment{
			"GLOBAL": "g",
		},
		Services: map[string]ServiceConfig{
			"svc": {
				"SVC_EXTRA": "s",
			},
		},
	}

	got := cfg.PerService("svc")
	want := orderedmap.NewOrderedMapWithElements(
		&orderedmap.Element[string, string]{Key: "GLOBAL", Value: "g"},
		&orderedmap.Element[string, string]{Key: "SVC_EXTRA", Value: "s"},
	)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ConfigPerService mismatch\nwant=%#v\ngot =%#v", want, got)
	}
}

func TestGetEnabledServices_FiltersCorrectly(t *testing.T) {
	cfg := Config{
		Environment: Environment{
			"GLOBAL": "g",
		},
		Services: map[string]ServiceConfig{
			"svc": {
				"SVC_EXTRA": "s",
			},
			"svc2": {
				"SVC_EXTRA": "s",
			},
		},
	}

	want := []string{"svc", "svc2"}
	assert.ElementsMatch(t, want, cfg.GetEnabledServices())
}

func TestObfuscateToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{"Empty token", "", ""},
		{"Short token", "123", strings.Repeat("*", 30)},
		{"Long token", "12345678901234567890", "1234567890" + strings.Repeat("*", 20)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Obfuscate(tt.token))
		})
	}
}

func TestIsObfuscated(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{"Obfuscated token", "1234567890********************", true},
		{"Not obfuscated", "1234567890", false},
		{"Empty token", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsObfuscated(tt.token))
		})
	}
}

func TestConfigGetBranch(t *testing.T) {
	t.Run("custom branch", func(t *testing.T) {
		cfg := Config{Settings: Settings{Git: GitConfig{Branch: "develop"}}}
		assert.Equal(t, "develop", cfg.GetBranch())
	})

	t.Run("empty branch defaults to main", func(t *testing.T) {
		cfg := Config{Settings: Settings{Git: GitConfig{Branch: ""}}}
		assert.Equal(t, DefaultBranch, cfg.GetBranch())
	})
}

func TestConfigIsEventNotificationEnabled(t *testing.T) {
	cfg := Config{
		Settings: Settings{
			Notifications: NotificationConfig{
				NotificationTypes: []EventType{EventDeploymentStarted, EventDeploymentError},
			},
		},
	}

	assert.True(t, cfg.IsEventNotificationEnabled(EventDeploymentStarted))
	assert.True(t, cfg.IsEventNotificationEnabled(EventDeploymentError))
	assert.False(t, cfg.IsEventNotificationEnabled(EventDeploymentSuccess))
}

func TestSettingsGetObfuscatedToken(t *testing.T) {
	t.Run("long token", func(t *testing.T) {
		s := Settings{Git: GitConfig{Token: "1234567890abcdef1234567890"}}
		result := s.GetObfuscatedToken()
		assert.Equal(t, "1234567890"+repeatAsterisks(20), result)
	})

	t.Run("short token", func(t *testing.T) {
		s := Settings{Git: GitConfig{Token: "short"}}
		result := s.GetObfuscatedToken()
		assert.Equal(t, repeatAsterisks(30), result)
	})

	t.Run("empty token", func(t *testing.T) {
		s := Settings{Git: GitConfig{Token: ""}}
		assert.Empty(t, s.GetObfuscatedToken())
	})
}

func TestSettingsGetObfuscatedNotificationURL(t *testing.T) {
	s := Settings{Notifications: NotificationConfig{NotificationURL: "gotify://localhost:8080"}}
	result := s.GetObfuscatedNotificationURL()
	assert.Equal(t, "gotify://l"+repeatAsterisks(20), result)
}

func TestOidcConfigGetObfuscatedClientSecret(t *testing.T) {
	t.Run("long secret", func(t *testing.T) {
		cfg := OidcConfig{ClientSecret: "1234567890abcdef1234567890"}
		result := cfg.GetObfuscatedClientSecret()
		assert.Equal(t, "1234567890"+repeatAsterisks(20), result)
	})

	t.Run("short secret", func(t *testing.T) {
		cfg := OidcConfig{ClientSecret: "short"}
		result := cfg.GetObfuscatedClientSecret()
		assert.Equal(t, repeatAsterisks(30), result)
	})

	t.Run("empty secret", func(t *testing.T) {
		cfg := OidcConfig{ClientSecret: ""}
		assert.Empty(t, cfg.GetObfuscatedClientSecret())
	})
}

func repeatAsterisks(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString("*")
	}
	return sb.String()
}
