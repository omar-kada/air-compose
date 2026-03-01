package mappers

import (
	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"
)

// FeaturesMapper maps models.Features to api.Features
type FeaturesMapper struct{}

// Map converts a models.Features to an api.Features
func (FeaturesMapper) Map(features models.Features) api.Features {
	return api.Features{
		DisplayConfig: features.DisplayConfig,
		EditConfig:    features.EditConfig,
		EditSettings:  features.EditSettings,
	}
}
