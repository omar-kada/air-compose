package models

import (
	"time"

	"github.com/moby/moby/api/types/container"
)

// StackStatus defines model for Stack.Status.
type StackStatus string

// Defines values for StackStatus.
const (
	StackStatusUnknown   StackStatus = "unknown"
	StackStatusUnhealthy StackStatus = "unhealthy"
	StackStatusStarting  StackStatus = "starting"
	StackStatusHealthy   StackStatus = "healthy"
)

// State defines model for State of AirCompose.
type State struct {
	LastStatus  DeploymentStatus
	NextDeploy  time.Time
	Health      StackStatus
	Initialized bool
}

// StacksState represents the state of multiple services in a stack.
type StacksState struct {
	services     map[string]StackStatus
	globalStatus StackStatus
}

// NewStacksState creates a new StacksState with empty services map and unknown global status.
func NewStacksState() StacksState {
	return StacksState{
		services:     make(map[string]StackStatus),
		globalStatus: StackStatusUnknown,
	}
}

// ForService returns the current status of the specified service.
// If the service is not found, it returns StackStatusUnknown.
func (ss *StacksState) ForService(serviceName string) StackStatus {
	if status, ok := ss.services[serviceName]; ok {
		return status
	}
	return StackStatusUnknown
}

// ProgressiveUpdateServiceStatus updates the status of a service and propagates the change to the global status.
func (ss *StacksState) ProgressiveUpdateServiceStatus(serviceName string, newStatus StackStatus) {
	ss.services[serviceName] = getCombinedStatus(ss.ForService(serviceName), newStatus)
	ss.globalStatus = getCombinedStatus(ss.globalStatus, newStatus)
}

// CombineContainerStatus updates the status of a service based on the container's health and state.
func (ss *StacksState) CombineContainerStatus(serviceName string, ctr ContainerSummary) {
	switch ctr.Health {
	case container.Healthy:
		ss.ProgressiveUpdateServiceStatus(serviceName, StackStatusHealthy)
	case container.Unhealthy:
		ss.ProgressiveUpdateServiceStatus(serviceName, StackStatusUnhealthy)
	case container.Starting:
		ss.ProgressiveUpdateServiceStatus(serviceName, StackStatusStarting)
	case container.NoHealthcheck:
		if ctr.State == container.StateRunning {
			ss.ProgressiveUpdateServiceStatus(serviceName, StackStatusHealthy)
		} else {
			ss.ProgressiveUpdateServiceStatus(serviceName, StackStatusUnhealthy)
		}
	}
}

// GetGlobalHealth returns the current global status of the stack.
func (ss *StacksState) GetGlobalHealth() StackStatus {
	return ss.globalStatus
}

func getCombinedStatus(oldStatus, newStatus StackStatus) StackStatus {
	switch oldStatus {
	case StackStatusUnhealthy:
		return oldStatus
	case StackStatusStarting:
		if newStatus != StackStatusUnhealthy {
			return oldStatus
		}
	case StackStatusHealthy:
		if newStatus == StackStatusUnknown {
			return oldStatus
		}
	}
	return newStatus
}
