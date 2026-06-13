package testutil

import "omar-kada/air-compose/internal/models"

// ConfigGetter provides a way to get and set configuration in tests without dependencies.
type ConfigGetter struct {
	config models.Config
}

// NewConfigGetter creates a new ConfigGetter with the given configuration.
func NewConfigGetter(config models.Config) *ConfigGetter {
	return &ConfigGetter{
		config: config,
	}
}

// Get returns the current configuration.
func (c *ConfigGetter) Get() models.Config {
	return c.config
}

// Set updates the configuration with the provided value.
func (c *ConfigGetter) Set(cfg models.Config) {
	c.config = cfg
}
