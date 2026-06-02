package docker

import (
	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/testutil/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInspector struct {
	mock.Mock
}

func (m *MockInspector) GetManagedStacks() (models.StacksState, error) {
	args := m.Called()
	return args.Get(0).(models.StacksState), args.Error(1)
}

func (m *MockInspector) GetCurrentStacks(services []string) (models.StacksState, error) {
	args := m.Called(services)
	return args.Get(0).(models.StacksState), args.Error(1)
}

func newCheckerWithMock(t *testing.T, inspector Inspector, mockDispatcher events.Dispatcher) *healthChecker {
	configStore, err := config.NewConfigStore(t.TempDir() + "/config.yaml")
	if err != nil {
		t.Fatal("error while creating config store", err)
	}
	err = configStore.Update(models.Config{
		Services: map[string]models.ServiceConfig{
			"service1": {},
		},
	})
	if err != nil {
		t.Fatal("error updating config", err)
	}
	return NewHealthChecker(configStore, inspector, mockDispatcher).(*healthChecker)
}

func TestStateRefresh_Table(t *testing.T) {
	tests := []struct {
		name          string
		initialState  models.ContainerHealth
		mockState     models.ContainerHealth
		expectedEvent models.EventType
	}{
		{
			name:          "Unhealthy to Healthy",
			initialState:  models.ContainerUnhealthy,
			mockState:     models.ContainerHealthy,
			expectedEvent: models.EventStacksHealthy,
		},
		{
			name:          "Healthy to Unhealthy",
			initialState:  models.ContainerHealthy,
			mockState:     models.ContainerUnhealthy,
			expectedEvent: models.EventStacksUnhealthy,
		},
		{
			name:          "Same Health",
			initialState:  models.ContainerHealthy,
			mockState:     models.ContainerHealthy,
			expectedEvent: "",
		},
		{
			name:          "First Run",
			initialState:  "",
			mockState:     models.ContainerHealthy,
			expectedEvent: "",
		},
		{
			name:          "First Run Unhealthy",
			initialState:  "",
			mockState:     models.ContainerUnhealthy,
			expectedEvent: models.EventStacksUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInspector := new(MockInspector)
			mockDispatcher := new(mocks.Dispatcher)
			healthChecker := newCheckerWithMock(t, mockInspector, mockDispatcher)

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

			// Set up dispatcher expectation if needed
			if tt.expectedEvent != "" {
				mockDispatcher.On("Dispatch", mock.Anything, tt.expectedEvent, mock.Anything).Once()
			}

			healthChecker.refreshState()

			if tt.expectedEvent == "" {
				mockDispatcher.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything, mock.Anything)
			}
			assert.Equal(t, tt.mockState, healthChecker.currentStacksState.GetGlobalHealth())
			mockDispatcher.AssertExpectations(t)
			mockInspector.AssertExpectations(t)
		})
	}
}
