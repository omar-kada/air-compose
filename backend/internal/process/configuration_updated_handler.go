package process

import (
	"context"
	"log/slog"
	"maps"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"reflect"
)

// ConfigurationUpdatedHandler handles configuration update events.
type ConfigurationUpdatedHandler struct {
	deploymentService DeploymentService
	eventPublisher    events.Publisher
}

// NewConfigurationUpdatedHandler creates a new ConfigurationUpdatedHandler with the given deployment service.
func NewConfigurationUpdatedHandler(deploymentService DeploymentService, eventPublisher events.Publisher) *ConfigurationUpdatedHandler {

	return &ConfigurationUpdatedHandler{
		deploymentService: deploymentService,
		eventPublisher:    eventPublisher,
	}
}

// HandleEvent processes configuration update events.
func (h *ConfigurationUpdatedHandler) HandleEvent(_ context.Context, event models.Event) {
	if event.Type == models.EventConfigurationUpdated {
		cfgChanges, ok := event.Data.(models.EventDataChange[models.Config])
		if ok {
			oldCfg := cfgChanges.Old
			newCfg := cfgChanges.New
			if maps.Equal(oldCfg.Environment, newCfg.Environment) &&
				reflect.DeepEqual(oldCfg.Services, newCfg.Services) &&
				reflect.DeepEqual(oldCfg.Settings.Git, newCfg.Settings.Git) {
				slog.Info("No relevant configuration changes detected, skipping deployment")
				return
			}
		} else {
			slog.Warn("configuration updated event has wrong data type", "event", event)
		}
		_, err := h.deploymentService.DoDeploy(DeploymentTriggerConfigurationUpdated, models.Patch{})
		if err != nil {
			slog.Error("error in deployment after configuration update", "err", err)
			h.eventPublisher.Publish(context.Background(), models.SourceEvent{
				Type: models.EventError,
				Msg:  "error in deployment after configuration update",
				Data: err,
			})
		}
	}
}
