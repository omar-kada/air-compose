package mappers

import (
	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"
)

// ConfigMapper maps models.Config to api.Config
type ConfigMapper struct{}

// Map converts a models.Config to an api.Config
func (ConfigMapper) Map(config models.Config) api.Config {
	convertedMap := make(map[string]map[string]string)

	for key, innerMap := range config.Services {
		convertedMap[key] = innerMap
	}

	return api.Config{
		GlobalVariables: config.Environment,
		Services:        convertedMap,
	}
}

// UnMap transforms back from api.Config to models.Config
func (ConfigMapper) UnMap(config api.Config) models.Config {
	convertedMap := make(map[string]models.ServiceConfig)

	for key, innerMap := range config.Services {
		convertedMap[key] = innerMap
	}

	return models.Config{
		Environment: config.GlobalVariables,
		Services:    convertedMap,
	}
}
