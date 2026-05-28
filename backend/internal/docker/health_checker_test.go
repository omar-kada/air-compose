package docker

import (
	"context"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDispatcher struct {
	mock.Mock
}

func (m *MockDispatcher) Dispatch(ctx context.Context, eventType models.EventType, message string) {
	m.Called(ctx, eventType, message)
}

type MockInspector struct {
	mock.Mock
}

func (m *MockInspector) GetStacksState() (models.StacksState, error) {
	args := m.Called()
	return args.Get(0).(models.StacksState), args.Error(1)
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
	configStore := storage.NewConfigStore(t.TempDir() + "/config.yaml")
	configStore.Update(models.Config{
		Services: map[string]models.ServiceConfig{
			"service1": {},
		},
	})
	return NewHealthChecker(configStore, inspector, mockDispatcher).(*healthChecker)
}

func TestStateRefresh_TriggersUnhealthyEvent(t *testing.T) {

	mockInspector := new(MockInspector)
	mockDispatcher := new(MockDispatcher)
	healthChecker := newCheckerWithMock(t, mockInspector, mockDispatcher)

	mockState := models.NewStacksState()
	mockState.SetContainerStatus("service1",
		models.ContainerSummary{
			ID:     "id",
			Name:   "container",
			Health: models.ContainerUnhealthy,
			State:  models.StateRunning,
		})
	mockInspector.On("GetCurrentStacks", mock.Anything).Return(
		mockState,
		nil,
	)
	// Expect dispatcher to be called with STACKS_UNHEALTHY
	mockDispatcher.On("Dispatch", mock.Anything, models.EventStacksUnhealthy, mock.Anything).Once()

	healthChecker.refreshState()

	assert.Equal(t, models.ContainerUnhealthy, healthChecker.currentStacksState.GetGlobalHealth())

	mockDispatcher.AssertExpectations(t)
	mockInspector.AssertExpectations(t)
}

func TestStateRefresh_DoesNotTriggerEventWhenHealthy(t *testing.T) {
	mockInspector := new(MockInspector)
	mockDispatcher := new(MockDispatcher)
	healthChecker := newCheckerWithMock(t, mockInspector, mockDispatcher)
	mockState := models.NewStacksState()
	mockState.SetContainerStatus("service1",
		models.ContainerSummary{
			ID:     "id",
			Name:   "container",
			Health: models.ContainerHealthy,
			State:  models.StateRunning,
		})
	mockInspector.On("GetCurrentStacks", mock.Anything).Return(
		mockState,
		nil,
	)
	healthChecker.refreshState()

	mockDispatcher.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything, mock.Anything)
	assert.Equal(t, models.ContainerHealthy, healthChecker.currentStacksState.GetGlobalHealth())
	mockInspector.AssertExpectations(t)
}
