package mappers

import (
	"testing"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
)

func TestStatsMapper_Map(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	next := now.Add(24 * time.Hour)

	cases := []struct {
		name string
		in   models.Stats
		want api.Stats
	}{
		{
			name: "basic",
			in: models.Stats{
				Author:     "alice",
				Error:      1,
				Success:    2,
				LastDeploy: now,
				NextDeploy: next,
				LastStatus: models.DeploymentStatusRunning,
				Health:     models.StackStatusHealthy,
			},
			want: api.Stats{
				Author:     "alice",
				Error:      1,
				Success:    2,
				LastDeploy: now,
				NextDeploy: next,
				Status:     api.DeploymentStatus(models.DeploymentStatusRunning),
				Health:     api.ContainerHealthHealthy,
			},
		},
		{
			name: "zero-times-empty-health",
			in: models.Stats{
				Author:     "bob",
				Error:      0,
				Success:    0,
				LastDeploy: time.Time{},
				NextDeploy: time.Time{},
				LastStatus: models.DeploymentStatusPlanned,
				Health:     models.StackStatusUnknown,
			},
			want: api.Stats{
				Author:     "bob",
				Error:      0,
				Success:    0,
				LastDeploy: time.Time{},
				NextDeploy: time.Time{},
				Status:     api.DeploymentStatus(models.DeploymentStatusPlanned),
				Health:     api.ContainerHealthUnknown,
			},
		},
		{
			name: "zero-times-empty-health",
			in: models.Stats{
				Author:     "foo",
				Error:      0,
				Success:    0,
				LastDeploy: time.Time{},
				NextDeploy: time.Time{},
				LastStatus: models.DeploymentStatusPlanned,
				Health:     models.StackStatusStarting,
			},
			want: api.Stats{
				Author:     "foo",
				Error:      0,
				Success:    0,
				LastDeploy: time.Time{},
				NextDeploy: time.Time{},
				Status:     api.DeploymentStatus(models.DeploymentStatusPlanned),
				Health:     api.ContainerHealthStarting,
			},
		},
		{
			name: "zero-times-empty-health",
			in: models.Stats{
				Author:     "bob",
				Error:      0,
				Success:    0,
				LastDeploy: time.Time{},
				NextDeploy: time.Time{},
				LastStatus: models.DeploymentStatusPlanned,
				Health:     models.StackStatusUnhealthy,
			},
			want: api.Stats{
				Author:     "bob",
				Error:      0,
				Success:    0,
				LastDeploy: time.Time{},
				NextDeploy: time.Time{},
				Status:     api.DeploymentStatus(models.DeploymentStatusPlanned),
				Health:     api.ContainerHealthUnhealthy,
			},
		},
	}

	m := StatsMapper{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.Map(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
