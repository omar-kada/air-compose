package mappers

import (
	"fmt"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"
)

// DeploymentMapper maps between models.Deployment and api.Deployment types.
type DeploymentMapper interface {
	Mapper[models.Deployment, api.Deployment]
	MapToPageInfo(deps []models.Deployment, limit int) api.PageInfo
}

type depMapper struct{}

// NewDeploymentMapper creates a new DeploymentMapper with the given DiffMapper and EventMapper.
func NewDeploymentMapper() DeploymentMapper {
	return depMapper{}
}

// Map maps a models.Deployment to an api.Deployment.
func (depMapper) Map(dep models.Deployment) api.Deployment {
	return api.Deployment{
		Author:  dep.Author,
		Diff:    dep.Diff,
		Id:      fmt.Sprintf("%d", dep.ID),
		Status:  api.DeploymentStatus(dep.Status),
		Time:    dep.Time,
		EndTime: dep.EndTime,
		Title:   dep.Title,
	}
}

// MapToPageInfo maps a slice of models.Deployment to an api.PageInfo, determining if there are more items
// and providing the end cursor for pagination.
func (depMapper) MapToPageInfo(deps []models.Deployment, limit int) api.PageInfo {
	return MapToPageInfo(deps, limit, func(dep models.Deployment) string {
		return fmt.Sprintf("%d", dep.ID)
	})
}
