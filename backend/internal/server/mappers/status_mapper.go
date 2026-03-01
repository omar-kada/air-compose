package mappers

import (
	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"
)

// StatusMapper maps models.ContainerSummary to api.ContainerStatus
type StatusMapper struct{}

// Map converts a models.ContainerSummary to an api.ContainerStatus
func (StatusMapper) Map(container models.ContainerSummary) api.ContainerStatus {
	return api.ContainerStatus{
		ContainerId: container.ID,
		State:       api.ContainerStatusState(container.State),
		Name:        container.Name,
		Health:      api.ContainerHealth(container.Health),
		StartedAt:   container.StartedAt,
	}
}
