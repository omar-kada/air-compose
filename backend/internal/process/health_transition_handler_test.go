package process

import (
	"context"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeploymentService is a mock implementation of DeploymentService
// for testing purposes.
type MockDeploymentService struct {
	mock.Mock
}

func (m *MockDeploymentService) DoDeploy(trigger DeploymentTrigger, patch models.Patch) (models.Deployment, error) {
	args := m.Called(trigger, patch)
	return args.Get(0).(models.Deployment), args.Error(1)
}

func TestHandleHealthCheck_Unhealthy(t *testing.T) {
	// Set up the mock config
	mockConfig := models.Config{
		Settings: models.Settings{
			Schedule: models.ScheduleConfig{
				RetriesOnUnhealthy: 3,
				RetryDelay:         2000,
			},
		},
	}
	// Create mocks
	configStore := testutil.NewConfigGetter(mockConfig)
	mockDeploymentService := new(MockDeploymentService)

	// Create the HealthTransitionHandler
	handler := NewHealthTransitionHandler(configStore, mockDeploymentService)

	// Set up the mock deployment service
	mockDeploymentService.On("DoDeploy", DeploymentTriggerUnhealthyStacks, models.Patch{}).Return(models.Deployment{}, nil)

	handler.HandleHealthCheck(context.Background(), models.Event{
		Type: models.EventHealthChange,
		Data: models.EventDataChange[models.ContainerHealth]{
			Old: models.ContainerHealthy,
			New: models.ContainerUnhealthy,
		},
	})
	time.Sleep(time.Millisecond)
	// Assert that the reset is incremented
	handler.mu.Lock()
	assert.Equal(t, 1, handler.retries)
	assert.NotNil(t, handler.retryTimer)
	handler.mu.Unlock()

	// should reset on healthy

	handler.HandleHealthCheck(context.Background(), models.Event{
		Type: models.EventHealthChange,
		Data: models.EventDataChange[models.ContainerHealth]{
			Old: models.ContainerUnhealthy,
			New: models.ContainerHealthy,
		},
	})
	handler.mu.Lock()
	assert.Equal(t, 0, handler.retries)
	assert.Nil(t, handler.retryTimer)
	handler.mu.Unlock()
	mockDeploymentService.AssertExpectations(t)
}

func TestHandleHealthCheck_StartingAfterUnhealthy(t *testing.T) {
	// Set up the mock config
	mockConfig := models.Config{
		Settings: models.Settings{
			Schedule: models.ScheduleConfig{
				RetriesOnUnhealthy: 3,
				RetryDelay:         5,
			},
		},
	}
	// Create mocks
	configStore := testutil.NewConfigGetter(mockConfig)
	mockDeploymentService := new(MockDeploymentService)

	// Create the HealthTransitionHandler
	handler := NewHealthTransitionHandler(configStore, mockDeploymentService)

	// Set up the mock deployment service
	mockDeploymentService.On("DoDeploy", DeploymentTriggerUnhealthyStacks, models.Patch{}).Return(models.Deployment{}, nil)

	handler.HandleHealthCheck(context.Background(), models.Event{
		Type: models.EventHealthChange,
		Data: models.EventDataChange[models.ContainerHealth]{
			Old: models.ContainerHealthy,
			New: models.ContainerUnhealthy,
		},
	})
	time.Sleep(time.Millisecond)
	// Assert that the reset is incremented
	handler.mu.Lock()
	assert.Equal(t, 1, handler.retries)
	assert.NotNil(t, handler.retryTimer)
	handler.mu.Unlock()

	// should reset on healthy

	handler.HandleHealthCheck(context.Background(), models.Event{
		Type: models.EventHealthChange,
		Data: models.EventDataChange[models.ContainerHealth]{
			Old: models.ContainerUnhealthy,
			New: models.ContainerStarting,
		},
	})
	time.Sleep(10 * time.Millisecond)
	handler.mu.Lock()
	assert.Equal(t, 1, handler.retries)
	assert.NotNil(t, handler.retryTimer)
	handler.mu.Unlock()
	mockDeploymentService.AssertExpectations(t)
}

func TestHandleHealthCheck_MaxRetriesReached(t *testing.T) {
	// Set up the mock config
	mockConfig := models.Config{
		Settings: models.Settings{
			Schedule: models.ScheduleConfig{
				RetriesOnUnhealthy: 2,
				RetryDelay:         1,
			},
		},
	}
	// Create mocks
	configStore := testutil.NewConfigGetter(mockConfig)
	mockDeploymentService := new(MockDeploymentService)

	// Create the HealthTransitionHandler
	handler := NewHealthTransitionHandler(configStore, mockDeploymentService)

	// Set up the mock deployment service
	mockDeploymentService.On("DoDeploy", DeploymentTriggerUnhealthyStacks, models.Patch{}).Return(models.Deployment{}, nil)

	// Send unhealthy health checks until max retries is reached
	handler.HandleHealthCheck(context.Background(), models.Event{
		Type: models.EventHealthChange,
		Data: models.EventDataChange[models.ContainerHealth]{
			Old: models.ContainerHealthy,
			New: models.ContainerUnhealthy,
		},
	})

	time.Sleep(10 * time.Millisecond)
	// Assert that the retries are incremented to max retries + 1
	handler.mu.Lock()
	defer handler.mu.Unlock()
	assert.Equal(t, 3, handler.retries)
	mockDeploymentService.AssertNumberOfCalls(t, "DoDeploy", 2)
}

func TestHandleHealthCheck_Healthy(t *testing.T) {
	// Create mocks
	configStore := testutil.NewConfigGetter(models.Config{})
	mockDeploymentService := new(MockDeploymentService)

	// Create the HealthTransitionHandler
	handler := NewHealthTransitionHandler(configStore, mockDeploymentService)

	// Set the current health to unhealthy
	handler.currentHealth = models.ContainerUnhealthy

	// Send a healthy health check
	handler.HandleHealthCheck(context.Background(), models.Event{
		Type: models.EventHealthChange,
		Data: models.EventDataChange[models.ContainerHealth]{
			Old: models.ContainerUnhealthy,
			New: models.ContainerHealthy,
		},
	})

	// Assert that the retries are reset
	handler.mu.Lock()
	defer handler.mu.Unlock()
	assert.Equal(t, 0, handler.retries)
	assert.Nil(t, handler.retryTimer)
}

func TestResetRetries(t *testing.T) {
	// Create mocks
	configStore := testutil.NewConfigGetter(models.Config{})
	mockDeploymentService := new(MockDeploymentService)

	// Create the HealthTransitionHandler
	handler := NewHealthTransitionHandler(configStore, mockDeploymentService)

	// Set the retries to a non-zero value
	handler.retries = 5

	// Reset the retries
	handler.ResetRetries()

	// Assert that the retries are reset
	assert.Equal(t, 0, handler.retries)
}
