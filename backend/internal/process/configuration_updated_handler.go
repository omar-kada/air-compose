package process

import (
	"context"
	"log/slog"
	"omar-kada/air-compose/internal/models"
)

// ConfigurationUpdatedHandler handles configuration update events.
type ConfigurationUpdatedHandler struct {
	deploymentService DeploymentService
}

// NewConfigurationUpdatedHandler creates a new ConfigurationUpdatedHandler with the given deployment service.
func NewConfigurationUpdatedHandler(deploymentService DeploymentService) *ConfigurationUpdatedHandler {

	return &ConfigurationUpdatedHandler{
		deploymentService: deploymentService,
	}
}

// HandleEvent processes configuration update events.
func (h *ConfigurationUpdatedHandler) HandleEvent(_ context.Context, event models.Event) {
	if event.Type == models.EventConfigurationUpdated {
		_, err := h.deploymentService.DoDeploy(DeploymentTriggerConfigurationUpdated, models.Patch{})
		if err != nil {
			slog.Error("error in deployment after configuration update", "err", err)
		}
	}
}
