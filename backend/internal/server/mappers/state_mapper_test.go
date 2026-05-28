package mappers

import (
	"testing"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
)

func TestStateMapper_Map(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	next := now.Add(24 * time.Hour)

	cases := []struct {
		name string
		in   models.State
		want api.State
	}{
		{
			name: "basic",
			in: models.State{
				NextDeploy: next,
				LastStatus: models.DeploymentStatusRunning,
				Health:     models.ContainerHealthy,
			},
			want: api.State{
				NextDeploy: next,
				Status:     api.DeploymentStatus(models.DeploymentStatusRunning),
				Health:     api.ContainerHealthHealthy,
			},
		},
		{
			name: "zero-times-empty-health",
			in: models.State{
				NextDeploy: time.Time{},
				LastStatus: models.DeploymentStatusPlanned,
				Health:     models.ContainerNoHealth,
			},
			want: api.State{
				NextDeploy: time.Time{},
				Status:     api.DeploymentStatus(models.DeploymentStatusPlanned),
				Health:     api.ContainerHealthNone,
			},
		},
		{
			name: "zero-times-empty-health",
			in: models.State{
				NextDeploy: time.Time{},
				LastStatus: models.DeploymentStatusPlanned,
				Health:     models.ContainerStarting,
			},
			want: api.State{
				NextDeploy: time.Time{},
				Status:     api.DeploymentStatus(models.DeploymentStatusPlanned),
				Health:     api.ContainerHealthStarting,
			},
		},
		{
			name: "zero-times-empty-health",
			in: models.State{
				NextDeploy: time.Time{},
				LastStatus: models.DeploymentStatusPlanned,
				Health:     models.ContainerUnhealthy,
			},
			want: api.State{
				NextDeploy: time.Time{},
				Status:     api.DeploymentStatus(models.DeploymentStatusPlanned),
				Health:     api.ContainerHealthUnhealthy,
			},
		},
	}

	m := StateMapper{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.Map(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
