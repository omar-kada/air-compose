package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStacksState(t *testing.T) {
	ss := NewStacksState()
	assert.NotNil(t, ss)
	assert.Empty(t, ss)
}

func TestSetContainerStatus(t *testing.T) {
	t.Run("new service", func(t *testing.T) {
		ss := NewStacksState()
		ctr := ContainerSummary{Name: "ctr1", Health: ContainerHealthy, State: StateRunning}
		ss.SetContainerStatus("svc1", ctr)
		assert.Len(t, ss, 1)
		assert.Len(t, ss["svc1"], 1)
		assert.Equal(t, ctr, ss["svc1"]["ctr1"])
	})

	t.Run("existing service adds container", func(t *testing.T) {
		ss := NewStacksState()
		ctr1 := ContainerSummary{Name: "ctr1", Health: ContainerHealthy}
		ctr2 := ContainerSummary{Name: "ctr2", Health: ContainerUnhealthy}
		ss.SetContainerStatus("svc1", ctr1)
		ss.SetContainerStatus("svc1", ctr2)
		assert.Len(t, ss, 1)
		assert.Len(t, ss["svc1"], 2)
	})
}

func TestGetUnhealthyServices(t *testing.T) {
	t.Run("mixed health", func(t *testing.T) {
		ss := NewStacksState()
		ss.SetContainerStatus("svc1", ContainerSummary{Name: "c1", Health: ContainerHealthy})
		ss.SetContainerStatus("svc2", ContainerSummary{Name: "c2", Health: ContainerUnhealthy})
		ss.SetContainerStatus("svc3", ContainerSummary{Name: "c3", Health: ContainerHealthy})
		ss.SetContainerStatus("svc3", ContainerSummary{Name: "c4", Health: ContainerUnhealthy})

		result := ss.GetUnhealthyServices()
		assert.ElementsMatch(t, []string{"svc2", "svc3"}, result)
	})

	t.Run("all healthy", func(t *testing.T) {
		ss := NewStacksState()
		ss.SetContainerStatus("svc1", ContainerSummary{Name: "c1", Health: ContainerHealthy})
		assert.Empty(t, ss.GetUnhealthyServices())
	})

	t.Run("empty", func(t *testing.T) {
		ss := NewStacksState()
		assert.Empty(t, ss.GetUnhealthyServices())
	})
}

func TestIsDeploying(t *testing.T) {
	t.Run("starting container", func(t *testing.T) {
		ss := NewStacksState()
		ss.SetContainerStatus("svc1", ContainerSummary{Health: ContainerStarting})
		assert.True(t, ss.IsDeploying())
	})

	t.Run("created state", func(t *testing.T) {
		ss := NewStacksState()
		ss.SetContainerStatus("svc1", ContainerSummary{State: StateCreated, Health: ContainerNoHealth})
		assert.True(t, ss.IsDeploying())
	})

	t.Run("restarting state", func(t *testing.T) {
		ss := NewStacksState()
		ss.SetContainerStatus("svc1", ContainerSummary{State: StateRestarting, Health: ContainerNoHealth})
		assert.True(t, ss.IsDeploying())
	})

	t.Run("healthy and running", func(t *testing.T) {
		ss := NewStacksState()
		ss.SetContainerStatus("svc1", ContainerSummary{Health: ContainerHealthy, State: StateRunning})
		assert.False(t, ss.IsDeploying())
	})

	t.Run("empty stacks", func(t *testing.T) {
		ss := NewStacksState()
		assert.False(t, ss.IsDeploying())
	})
}

func TestGetGlobalHealth(t *testing.T) {
	t.Run("empty stacks returns NoHealth", func(t *testing.T) {
		ss := NewStacksState()
		assert.Equal(t, ContainerNoHealth, ss.GetGlobalHealth())
	})

	t.Run("healthy containers", func(t *testing.T) {
		ss := NewStacksState()
		ss.SetContainerStatus("svc1", ContainerSummary{Health: ContainerHealthy})
		assert.Equal(t, ContainerHealthy, ss.GetGlobalHealth())
	})

	t.Run("unhealthy dominates", func(t *testing.T) {
		ss := NewStacksState()
		ss.SetContainerStatus("svc1", ContainerSummary{Health: ContainerHealthy})
		ss.SetContainerStatus("svc2", ContainerSummary{Health: ContainerUnhealthy})
		assert.Equal(t, ContainerUnhealthy, ss.GetGlobalHealth())
	})

	t.Run("starting dominates healthy", func(t *testing.T) {
		ss := NewStacksState()
		ss.SetContainerStatus("svc1", ContainerSummary{Health: ContainerHealthy})
		ss.SetContainerStatus("svc2", ContainerSummary{Health: ContainerStarting})
		assert.Equal(t, ContainerStarting, ss.GetGlobalHealth())
	})

	t.Run("empty service gets unhealthy", func(t *testing.T) {
		ss := StacksState{"svc1": {}}
		assert.Equal(t, ContainerUnhealthy, ss.GetGlobalHealth())
	})
}

func TestGetCombinedHealth(t *testing.T) {
	tests := []struct {
		name     string
		old      ContainerHealth
		new      ContainerHealth
		expected ContainerHealth
	}{
		{"old unhealthy", ContainerUnhealthy, ContainerHealthy, ContainerUnhealthy},
		{"new unhealthy", ContainerHealthy, ContainerUnhealthy, ContainerUnhealthy},
		{"old starting", ContainerStarting, ContainerHealthy, ContainerStarting},
		{"new starting", ContainerHealthy, ContainerStarting, ContainerStarting},
		{"both healthy", ContainerHealthy, ContainerHealthy, ContainerHealthy},
		{"both no health", ContainerNoHealth, ContainerNoHealth, ContainerNoHealth},
		{"old nohealth, new unhealthy", ContainerNoHealth, ContainerUnhealthy, ContainerUnhealthy},
		{"old nohealth, new starting", ContainerNoHealth, ContainerStarting, ContainerStarting},
		{"old nohealth, new healthy", ContainerNoHealth, ContainerHealthy, ContainerHealthy},
		{"unknown values fall through", ContainerHealth("weird"), ContainerHealth("weird2"), ContainerHealth("weird2")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getCombinedHealth(tt.old, tt.new))
		})
	}
}
