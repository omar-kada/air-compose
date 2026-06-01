package process

import (
	"log/slog"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/storage"
)

// HealthTransitionHandler processes container health transitions and handles redeployment
// when containers become unhealthy.
type HealthTransitionHandler struct {
	configStore       storage.ConfigStore
	deploymentService DeploymentService

	retries       int
	currentHealth models.ContainerHealth
}

// NewHealthTransitionHandler creates a new HealthTransitionHandler that processes container health transitions.
func NewHealthTransitionHandler(configStore storage.ConfigStore,
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
	switch health {
	case models.ContainerUnhealthy:
		cfg, err := h.configStore.Get()
		if err != nil {
			slog.Error(err.Error())
			return
		}

		if h.retries == cfg.Settings.Schedule.RetriesOnUnhealthy {
			slog.Error("max retries on unhealthy reached", "retries", h.retries)
			h.retries = h.retries + 1
			return
		} else if h.retries > cfg.Settings.Schedule.RetriesOnUnhealthy {
			return
		}

		h.retries = h.retries + 1
		slog.Debug("incrementing number of retries", "retries", h.retries)

		_, err = h.deploymentService.SyncDeployment()
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
	h.retries = 0
}
