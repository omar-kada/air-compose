package mappers

import (
	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"
)

// StatsMapper maps models.Stats to api.Stats
type StatsMapper struct{}

// Map converts a models.Stats to an api.Stats
func (StatsMapper) Map(stats models.Stats) api.Stats {
	return api.Stats{
		Author:     stats.Author,
		Error:      stats.Error,
		Success:    stats.Success,
		LastDeploy: stats.LastDeploy,
		NextDeploy: stats.NextDeploy,
		Status:     api.DeploymentStatus(stats.LastStatus),
		Health:     api.ContainerHealth(stats.Health),
	}
}
