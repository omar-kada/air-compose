package process

import (
	"log/slog"
	"omar-kada/air-compose/internal/models"
	"sync"
)

// HealthTransitionHandler processes container health transitions and handles redeployment
// when containers become unhealthy.
type HealthTransitionHandler struct {
	configStore       models.ConfigGetter
	deploymentService DeploymentService

	retries       int
	mu            sync.Mutex
	currentHealth models.ContainerHealth
}

// NewHealthTransitionHandler creates a new HealthTransitionHandler that processes container health transitions.
func NewHealthTransitionHandler(configStore models.ConfigGetter,
	deploymentService DeploymentService,
	healthCheckChan <-chan models.ContainerHealth) *HealthTransitionHandler {

	healthHandler := &HealthTransitionHandler{
		configStore:       configStore,
		deploymentService: deploymentService,
	}

	go func() {
		for health := range healthCheckChan {
			healthHandler.handleHealthCheck(health)
		}
	}()

	return healthHandler
}

func (h *HealthTransitionHandler) handleHealthCheck(health models.ContainerHealth) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch health {
	case models.ContainerUnhealthy:

		cfg := h.configStore.Get()

		if h.retries == cfg.Settings.Schedule.RetriesOnUnhealthy {
			slog.Error("max retries on unhealthy reached", "retries", h.retries)
			h.retries = h.retries + 1
			return
		} else if h.retries > cfg.Settings.Schedule.RetriesOnUnhealthy {
			return
		}

		h.retries = h.retries + 1
		slog.Debug("incrementing number of retries", "retries", h.retries)

		_, err := h.deploymentService.DoDeploy(DeploymentTriggerUnhealthyStacks, models.Patch{})
		if err != nil {
			slog.Error(err.Error())
			return
		}
	case models.ContainerHealthy:
		if h.currentHealth == models.ContainerHealthy {
			return
		}
		h.retries = 0
		slog.Debug("resetting number of retries to 0")
	default:
		return
	}
}

// ResetRetries resets the retry counter to zero.
func (h *HealthTransitionHandler) ResetRetries() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.retries = 0
}
