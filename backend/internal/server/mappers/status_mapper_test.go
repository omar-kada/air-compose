package mappers

import (
	"testing"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestStatusMapper_Map(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	cases := []struct {
		name string
		in   models.ContainerSummary
		want api.ContainerStatus
	}{
		{
			name: "running-healthy",
			in: models.ContainerSummary{
				ID:        "cid1",
				Name:      "c1",
				State:     container.ContainerState("running"),
				Health:    container.HealthStatus("healthy"),
				StartedAt: now,
			},
			want: api.ContainerStatus{
				ContainerId: "cid1",
				Name:        "c1",
				State:       api.ContainerStatusState("running"),
				Health:      api.ContainerHealth("healthy"),
				StartedAt:   now,
			},
		},
		{
			name: "exited-none",
			in: models.ContainerSummary{
				ID:        "cid2",
				Name:      "c2",
				State:     container.ContainerState("exited"),
				Health:    container.HealthStatus("none"),
				StartedAt: time.Time{},
			},
			want: api.ContainerStatus{
				ContainerId: "cid2",
				Name:        "c2",
				State:       api.ContainerStatusState("exited"),
				Health:      api.ContainerHealth("none"),
				StartedAt:   time.Time{},
			},
		},
	}

	m := StatusMapper{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.Map(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
