package mappers

import (
	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"
)

// StateMapper maps models.State to api.State
type StateMapper struct{}

// Map converts a models.State to an api.State
func (StateMapper) Map(state models.State) api.State {
	return api.State{
		NextDeploy:  state.NextDeploy,
		Status:      api.DeploymentStatus(state.LastStatus),
		Health:      api.ContainerHealth(state.Health),
		Initialized: state.Initialized,
	}
}
