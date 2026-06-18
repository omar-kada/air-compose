package process

import (
	"context"
	"errors"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"testing"

	"github.com/stretchr/testify/mock"
)

// MockProcess is a mock implementation of DeploymentService
// for testing purposes.
type MockProcess struct {
	mock.Mock
}

func (m *MockProcess) DoDeploy(trigger DeploymentTrigger, patch models.Patch) (models.Deployment, error) {
	args := m.Called(trigger, patch)
	return args.Get(0).(models.Deployment), args.Error(1)
}

func TestHandleEvent_ConfigurationUpdated(t *testing.T) {
	// Create mocks
	mockDeploymentService := new(MockProcess)

	// Create the ConfigurationUpdatedHandler
	handler := NewConfigurationUpdatedHandler(mockDeploymentService, events.NewBus(1))

	// Set up the mock deployment service
	mockDeploymentService.On("DoDeploy", DeploymentTriggerConfigurationUpdated, models.Patch{}).Return(models.Deployment{}, nil)

	// Create a configuration updated event
	event := models.Event{
		Type: models.EventConfigurationUpdated,
		Data: models.EventDataChange[models.Config]{
			Old: models.Config{Settings: models.Settings{Git: models.GitConfig{Repo: "old-repo"}}},
			New: models.Config{Settings: models.Settings{Git: models.GitConfig{Repo: "new-repo"}}},
		},
	}

	// Handle the event
	handler.HandleEvent(context.Background(), event)

	// Assert that the deployment service was called
	mockDeploymentService.AssertExpectations(t)
}

func TestHandleEvent_ConfigurationUpdated_NoChanges(t *testing.T) {
	// Create mocks
	mockDeploymentService := new(MockProcess)

	// Create the ConfigurationUpdatedHandler
	handler := NewConfigurationUpdatedHandler(mockDeploymentService, events.NewBus(1))

	// Set up the mock deployment service to not expect any call
	mockDeploymentService.AssertNotCalled(t, "DoDeploy")

	// Create a configuration updated event with no changes
	event := models.Event{
		Type: models.EventConfigurationUpdated,
		Data: models.EventDataChange[models.Config]{
			Old: models.Config{Settings: models.Settings{Git: models.GitConfig{Repo: "same-repo"}}},
			New: models.Config{Settings: models.Settings{Git: models.GitConfig{Repo: "same-repo"}}},
		},
	}

	// Handle the event
	handler.HandleEvent(context.Background(), event)

	// Assert that the deployment service was not called
	mockDeploymentService.AssertExpectations(t)
}

func TestHandleEvent_ConfigurationUpdated_Error(t *testing.T) {
	// Create mocks
	mockDeploymentService := new(MockProcess)

	// Create the ConfigurationUpdatedHandler
	handler := NewConfigurationUpdatedHandler(mockDeploymentService, events.NewBus(1))

	// Set up the mock deployment service to return an error
	mockDeploymentService.On("DoDeploy", DeploymentTriggerConfigurationUpdated, models.Patch{}).Return(models.Deployment{}, errors.New("deployment error"))

	// Create a configuration updated event
	event := models.Event{
		Type: models.EventConfigurationUpdated,
	}

	// Handle the event
	handler.HandleEvent(context.Background(), event)

	// Assert that the deployment service was called
	mockDeploymentService.AssertExpectations(t)
}
