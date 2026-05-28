package models

import (
	"time"

	"github.com/moby/moby/api/types/container"
)

// ContainerHealth defines the health of a container
type ContainerHealth container.HealthStatus

// Defines values for ContainerHealth.
const (
	ContainerNoHealth  ContainerHealth = ContainerHealth(container.NoHealthcheck)
	ContainerUnhealthy ContainerHealth = ContainerHealth(container.Unhealthy)
	ContainerStarting  ContainerHealth = ContainerHealth(container.Starting)
	ContainerHealthy   ContainerHealth = ContainerHealth(container.Healthy)
)

// ContainerState represents the state of a container.
type ContainerState container.ContainerState

// Defines values for ContainerState.
const (
	StateCreated    ContainerState = ContainerState(container.StateCreated)
	StateRunning    ContainerState = ContainerState(container.StateRunning)
	StateRestarting ContainerState = ContainerState(container.StateRestarting)
	StatePaused     ContainerState = ContainerState(container.StatePaused)
	StateRemoving   ContainerState = ContainerState(container.StateRemoving)
	StateExited     ContainerState = ContainerState(container.StateExited)
	StateDead       ContainerState = ContainerState(container.StateDead)
)

// ContainerStatus represents the health and state of a container.
type ContainerStatus struct {
	Health ContainerHealth
	State  ContainerState
}

// State defines model for State of AirCompose.
type State struct {
	LastStatus  DeploymentStatus
	NextDeploy  time.Time
	Health      ContainerHealth
	Initialized bool
}

// StacksState represents the state of multiple services in a stack.
type StacksState map[string]map[string]ContainerSummary

// NewStacksState creates a new StacksState with empty services map and unknown global status.
func NewStacksState() StacksState {
	return make(StacksState)
}

// SetContainerStatus updates the status of a container
func (ss StacksState) SetContainerStatus(serviceName string, ctr ContainerSummary) {
	_, found := ss[serviceName]
	if !found {
		ss[serviceName] = make(map[string]ContainerSummary)
	}
	ss[serviceName][ctr.Name] = ctr
}

// GetUnhealthyServices returns the list of services that are unhealthy.
func (ss StacksState) GetUnhealthyServices() []string {
	var unhealthy []string

	for service, status := range ss {
		for _, container := range status {
			if container.Health == ContainerUnhealthy {
				unhealthy = append(unhealthy, service)
				break
			}
		}
	}
	return unhealthy
}

// IsDeploying returns true if the stack is in a deploying state (starting, created, or restarting)
func (ss StacksState) IsDeploying() bool {
	for _, service := range ss {
		for _, container := range service {
			if container.Health == ContainerStarting || container.State == StateCreated || container.State == StateRestarting {
				return true
			}
		}
	}
	return false
}

// GetGlobalHealth returns the current global status of the stack.
func (ss StacksState) GetGlobalHealth() ContainerHealth {
	health := ContainerNoHealth
	for _, service := range ss {
		if len(service) == 0 {
			health = getCombinedHealth(health, ContainerNoHealth)
		}
		for _, container := range service {
			health = getCombinedHealth(health, container.Health)
		}
	}
	return health
}

var healthOrder = []ContainerHealth{
	ContainerUnhealthy, ContainerStarting, ContainerHealthy, ContainerNoHealth,
}

func getCombinedHealth(oldStatus, newStatus ContainerHealth) ContainerHealth {

	for _, state := range healthOrder {
		if oldStatus == state || newStatus == state {
			return state
		}
	}
	return newStatus
}
