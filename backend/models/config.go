package models

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/elliotchance/orderedmap/v3"
)

// DefaultBranch is the default branch name used when no branch is specified in the configuration.
const DefaultBranch = "main"

// Settings represents configuration of air-compose.
type Settings struct {
	Repo              string      `mapstructure:"repo"`
	Branch            string      `mapstructure:"branch"`
	Username          string      `mapstructure:"username"`
	Token             string      `mapstructure:"token"`
	Cron              string      `mapstructure:"cron"`
	NotificationURL   string      `mapstructure:"notificationURL"`
	NotificationTypes []EventType `mapstructure:"notificationTypes"`
}

// Environment represents global environment variables.
type Environment map[string]string

// ServiceConfig represents configuration for an individual service.
type ServiceConfig map[string]string

// Config represents the overall configuration structure.
type Config struct {
	Settings    Settings                 `mapstructure:"settings"`
	Environment Environment              `mapstructure:"environment"`
	Services    map[string]ServiceConfig `mapstructure:"services"`
}

// PerService generates a slice of configuration variables for a specific service
func (cfg Config) PerService(service string) *orderedmap.OrderedMap[string, string] {
	serviceConfig := orderedmap.NewOrderedMap[string, string]()

	for key, value := range cfg.Environment {
		serviceConfig.Set(strings.ToUpper(key), fmt.Sprint(value))
	}
	if svcVars, ok := cfg.Services[service]; ok {
		for key, value := range svcVars {
			serviceConfig.Set(strings.ToUpper(key), fmt.Sprint(value))
		}
	}
	return serviceConfig
}

// GetEnabledServices returns the list of enabled services on the configuration
func (cfg Config) GetEnabledServices() []string {
	return slices.Collect(maps.Keys(cfg.Services))
}

// GetBranch returns the branch name from the configuration. If no branch is specified,
// it defaults to "main".
func (cfg Config) GetBranch() string {
	if cfg.Settings.Branch != "" {
		return cfg.Settings.Branch
	}
	return DefaultBranch
}

// IsEventNotificationEnabled checks if the specified event type is enabled for notifications.
func (cfg Config) IsEventNotificationEnabled(eventType EventType) bool {
	return slices.Contains(cfg.Settings.NotificationTypes, eventType)
}

// GetObfuscatedToken returns an obfuscated token
func (settings Settings) GetObfuscatedToken() string {
	return Obfuscate(settings.Token)
}

// GetObfuscatedNotificationURL returns an obfuscated notification URL
func (settings Settings) GetObfuscatedNotificationURL() string {
	return Obfuscate(settings.NotificationURL)
}

// Obfuscate replaces most of the input with asterisks to hide sensitive information
func Obfuscate(token string) string {
	if token == "" {
		return token
	}
	length := len(token)
	if length < 20 {
		return strings.Repeat("*", 30)
	}
	return token[0:10] + strings.Repeat("*", 20)
}

// IsObfuscated checks if the token is obfuscated by checking if it starts with "*****".
func IsObfuscated(token string) bool {
	return strings.HasSuffix(token, strings.Repeat("*", 20))
}
