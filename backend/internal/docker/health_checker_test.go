package docker

import (
	"context"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/testutil"
	"omar-kada/air-compose/testutil/mocks"
	"testing"
	"time"

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
			mockEventPublisher := new(mocks.EventPublisher)
			healthChecker := newCheckerWithMock(t, mockInspector, mockEventPublisher)

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
			if tt.expectedEvent != "" {
				mockEventPublisher.On("Publish", mock.Anything,
					mock.MatchedBy(func(srcEvent models.SourceEvent) bool {
						return srcEvent.Type == tt.expectedEvent
					}),
				).Once()
			}

			healthChecker.refreshState()

			if tt.expectedEvent == "" {
				mockEventPublisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
			}
			assert.Equal(t, tt.mockState, healthChecker.currentStacksState.GetGlobalHealth())
			mockEventPublisher.AssertExpectations(t)
			mockInspector.AssertExpectations(t)
		})
	}
}

func TestScheduleStateRefresh(t *testing.T) {
	mockInspector := new(MockInspector)
	mockEventPublisher := new(mocks.EventPublisher)
	healthChecker := newCheckerWithMock(t, mockInspector, mockEventPublisher)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Mock inspector response
	mockState := models.NewStacksState()
	mockState.SetContainerStatus("service1",
		models.ContainerSummary{
			ID:     "id",
			Name:   "container",
			Health: models.ContainerHealthy,
			State:  models.StateRunning,
		})
	mockInspector.On("GetCurrentStacks", mock.Anything).Return(mockState, nil)
	healthChecker.refreshDuration = 10 * time.Millisecond
	// Start the refresh scheduler
	go healthChecker.ScheduleStateRefresh(ctx)

	// Wait for at least one refresh to occur
	time.Sleep(20 * time.Millisecond)
	cancel()
	// Verify the state was refreshed
	healthChecker.refreshMu.Lock()
	assert.Equal(t, models.ContainerHealthy, healthChecker.currentStacksState.GetGlobalHealth())
	healthChecker.refreshMu.Unlock()

	// Verify the channel was updated
	select {
	case health := <-healthChecker.GetChannel():
		assert.Equal(t, models.ContainerHealthy, health)
	default:
		t.Error("Expected health check channel to receive a value")
	}

	// Test cancellation
	cancel()
	time.Sleep(100 * time.Millisecond)
	mockInspector.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}
