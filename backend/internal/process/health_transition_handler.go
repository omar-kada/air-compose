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

		if !cfg.Settings.Schedule.RedeployOnUnhealthy || h.retries >= cfg.Settings.Schedule.MaxRetries {
			slog.Debug("couldn't retry",
				"redeployOnUnhealthy", cfg.Settings.Schedule.RedeployOnUnhealthy,
				"maxRetries", cfg.Settings.Schedule.MaxRetries)
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
