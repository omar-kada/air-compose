package docker

import (
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/testutil"
	"omar-kada/air-compose/testutil/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newCheckerWithMock(t *testing.T, inspector Inspector, mockEventPublisher events.Publisher) *healthChecker {
	t.Helper()
	configStore := testutil.NewConfigGetter(models.Config{
		Services: map[string]models.ServiceConfig{
			"service1": {},
		},
	})
	return NewHealthChecker(configStore, inspector, mockEventPublisher).(*healthChecker)
}

func TestStateRefresh_Table(t *testing.T) {
	tests := []struct {
		name          string
		initialState  models.ContainerHealth
		mockState     models.ContainerHealth
		expectedEvent *models.SourceEvent
	}{
		{
			name:          "Unhealthy to Healthy",
			initialState:  models.ContainerUnhealthy,
			mockState:     models.ContainerHealthy,
			expectedEvent: new(models.NewHealthChangedEvent(models.ContainerUnhealthy, models.ContainerHealthy, nil)),
		},
		{
			name:          "Healthy to Unhealthy",
			initialState:  models.ContainerHealthy,
			mockState:     models.ContainerUnhealthy,
			expectedEvent: new(models.NewHealthChangedEvent(models.ContainerHealthy, models.ContainerUnhealthy, []string{"service1"})),
		},
		{
			name:          "Same Health",
			initialState:  models.ContainerHealthy,
			mockState:     models.ContainerHealthy,
			expectedEvent: nil,
		},
		{
			name:          "First Run",
			initialState:  models.ContainerNoHealth,
			mockState:     models.ContainerHealthy,
			expectedEvent: new(models.NewHealthChangedEvent(models.ContainerNoHealth, models.ContainerHealthy, nil)),
		},
		{
			name:          "First Run Unhealthy",
			initialState:  "",
			mockState:     models.ContainerUnhealthy,
			expectedEvent: new(models.NewHealthChangedEvent(models.ContainerNoHealth, models.ContainerUnhealthy, []string{"service1"})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInspector := new(mocks.Inspector)
			mockEventPublisher := new(mocks.EventPublisher)
			healthChecker := newCheckerWithMock(t, mockInspector, mockEventPublisher)
			healthChecker.currentStacksState = models.NewStacksState()

			// Set initial state if needed
			if tt.initialState != "" {
				initialState := models.NewStacksState()
				initialState.SetContainerStatus("service1",
					models.ContainerSummary{
						ID:     "id",
						Name:   "container",
						Health: tt.initialState,
						State:  models.StateRunning,
					})
				healthChecker.currentStacksState = initialState
			}

			// Mock inspector response
			mockState := models.NewStacksState()
			mockState.SetContainerStatus("service1",
				models.ContainerSummary{
					ID:     "id",
					Name:   "container",
					Health: tt.mockState,
					State:  models.StateRunning,
				})
			mockInspector.On("GetCurrentStacks", mock.Anything).Return(
				mockState,
				nil,
			)

			// Set up eventPublisher expectation if needed
			if tt.expectedEvent != nil {
				mockEventPublisher.On("Publish", mock.Anything,
					*tt.expectedEvent,
				).Once()
			}

			healthChecker.refreshState()

			if tt.expectedEvent == nil {
				mockEventPublisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
			}
			assert.Equal(t, tt.mockState, healthChecker.currentStacksState.GetGlobalHealth())
			mockEventPublisher.AssertExpectations(t)
			mockInspector.AssertExpectations(t)
		})
	}
}
