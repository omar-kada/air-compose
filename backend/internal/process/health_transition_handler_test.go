package process

import (
	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeploymentService is a mock implementation of DeploymentService
// for testing purposes.
type MockDeploymentService struct {
	mock.Mock
	DeploymentService
}

// SyncDeployment is a mock implementation of the SyncDeployment method.
func (m *MockDeploymentService) SyncDeployment() (models.Deployment, error) {
	args := m.Called()
	return args.Get(0).(models.Deployment), args.Error(1)
}

func newConfigStore(t *testing.T, cfg models.Config) config.Store {
	configStore, err := config.NewConfigStore(t.TempDir() + "/config.yaml")
	if err != nil {
		t.Fatal("error while creating configStore", err)
	}
	err = configStore.Update(cfg)
	if err != nil {
		t.Fatal("error updating config", err)
	}
	return configStore
}

func TestHandleHealthCheck_Unhealthy(t *testing.T) {
	// Set up the mock config
	mockConfig := models.Config{
		Settings: models.Settings{
			Schedule: models.ScheduleConfig{
				RetriesOnUnhealthy: 3,
			},
		},
	}
	// Create mocks
	configStore := newConfigStore(t, mockConfig)
	mockDeploymentService := new(MockDeploymentService)
	// Create a channel for health checks
	healthCheckChan := make(chan models.ContainerHealth)

	// Create the HealthTransitionHandler
	handler := NewHealthTransitionHandler(configStore, mockDeploymentService, healthCheckChan)

	// Set up the mock deployment service
	mockDeploymentService.On("SyncDeployment").Return(models.Deployment{}, nil)

	// Send an unhealthy health check
	healthCheckChan <- models.ContainerUnhealthy

	// Close the channel
	close(healthCheckChan)

	// Assert that the reset is incremented
	handler.mu.Lock()
	defer handler.mu.Unlock()
	assert.Equal(t, 1, handler.retries)
	mockDeploymentService.AssertExpectations(t)
}

func TestHandleHealthCheck_MaxRetriesReached(t *testing.T) {
	// Set up the mock config
	mockConfig := models.Config{
		Settings: models.Settings{
			Schedule: models.ScheduleConfig{
				RetriesOnUnhealthy: 2,
			},
		},
	}
	// Create mocks
	configStore := newConfigStore(t, mockConfig)
	mockDeploymentService := new(MockDeploymentService)

	// Create a channel for health checks
	healthCheckChan := make(chan models.ContainerHealth)

	// Create the HealthTransitionHandler
	handler := NewHealthTransitionHandler(configStore, mockDeploymentService, healthCheckChan)

	// Set up the mock deployment service
	mockDeploymentService.On("SyncDeployment").Return(models.Deployment{}, nil)

	// Send unhealthy health checks until max retries is reached
	healthCheckChan <- models.ContainerUnhealthy
	healthCheckChan <- models.ContainerUnhealthy
	healthCheckChan <- models.ContainerUnhealthy

	// Close the channel
	close(healthCheckChan)

	time.Sleep(10 * time.Millisecond)
	// Assert that the retries are incremented to max retries + 1
	handler.mu.Lock()
	defer handler.mu.Unlock()
	assert.Equal(t, 3, handler.retries)
	mockDeploymentService.AssertNumberOfCalls(t, "SyncDeployment", 2)
}

func TestHandleHealthCheck_Healthy(t *testing.T) {
	// Create mocks
	configStore := newConfigStore(t, models.Config{})
	mockDeploymentService := new(MockDeploymentService)

	// Create a channel for health checks
	healthCheckChan := make(chan models.ContainerHealth)

	// Create the HealthTransitionHandler
	handler := NewHealthTransitionHandler(configStore, mockDeploymentService, healthCheckChan)

	// Set the current health to unhealthy
	handler.currentHealth = models.ContainerUnhealthy

	// Send a healthy health check
	healthCheckChan <- models.ContainerHealthy

	// Close the channel
	close(healthCheckChan)
	// Assert that the retries are reset
	handler.mu.Lock()
	defer handler.mu.Unlock()
	assert.Equal(t, 0, handler.retries)
}

func TestResetRetries(t *testing.T) {
	// Create mocks
	configStore := newConfigStore(t, models.Config{})
	mockDeploymentService := new(MockDeploymentService)

	// Create a channel for health checks
	healthCheckChan := make(chan models.ContainerHealth)

	// Create the HealthTransitionHandler
	handler := NewHealthTransitionHandler(configStore, mockDeploymentService, healthCheckChan)

	// Set the retries to a non-zero value
	handler.retries = 5

	// Reset the retries
	handler.ResetRetries()

	// Assert that the retries are reset
	assert.Equal(t, 0, handler.retries)

	// Close the channel
	close(healthCheckChan)
}
