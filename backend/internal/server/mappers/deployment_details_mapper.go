// Package mappers provides functionality for mapping between different data models.
package mappers

import (
	"fmt"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"
)

// DeploymentDetailsMapper maps between models.Deployment and api.DeploymentWithDetails types.
type DeploymentDetailsMapper interface {
	Mapper[models.Deployment, api.DeploymentWithDetails]
}

type depDetailsMapper struct {
	diffMapper  Mapper[models.FileDiff, api.FileDiff]
	eventMapper Mapper[models.Event, api.Event]
}

// NewDeploymentDetailsMapper creates a new DeploymentMapper with the given DiffMapper and EventMapper.
func NewDeploymentDetailsMapper(diffMapper Mapper[models.FileDiff, api.FileDiff], eventMapper Mapper[models.Event, api.Event]) DeploymentDetailsMapper {
	return depDetailsMapper{
		diffMapper:  diffMapper,
		eventMapper: eventMapper,
	}
}

// Map maps a models.Deployment to an api.DeploymentWithDetails.
func (m depDetailsMapper) Map(dep models.Deployment) api.DeploymentWithDetails {
	return api.DeploymentWithDetails{
		Author:  dep.Author,
		Diff:    dep.Diff,
		Id:      fmt.Sprintf("%d", dep.ID),
		Status:  api.DeploymentStatus(dep.Status),
		Time:    dep.Time,
		EndTime: dep.EndTime,
		Title:   dep.Title,
		Events:  models.ListMapper(m.eventMapper.Map)(dep.Events),
		Files:   models.ListMapper(m.diffMapper.Map)(dep.Files),
	}
}
