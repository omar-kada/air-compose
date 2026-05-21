package docker

import (
	"context"
	"errors"
	"testing"
	"time"

	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the Client interface
type MockClient struct {
	mock.Mock
}

func (m *MockClient) ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(client.ContainerListResult), args.Error(1)
}

func (m *MockClient) ContainerInspect(ctx context.Context, containerID string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	args := m.Called(ctx, containerID, options)
	return args.Get(0).(client.ContainerInspectResult), args.Error(1)
}

type MockExec struct {
	mock.Mock
}

func (m *MockExec) Exec(cmd string, cmdArgs ...string) ([]byte, error) {
	args := m.Called(cmd, cmdArgs)
	return args.Get(0).([]byte), args.Error(1)
}

type MockDispatcher struct {
	mock.Mock
}

func (m *MockDispatcher) Dispatch(ctx context.Context, eventType models.EventType, message string) {
	m.Called(ctx, eventType, message)
}

func newInspectorWithMock(t *testing.T, client Client, mockExec shell.Executor, mockDispatcher events.Dispatcher, servicesDir string) *inspector {
	configStore := storage.NewConfigStore(t.TempDir() + "/config.yaml")
	configStore.Update(models.Config{
		Services: map[string]models.ServiceConfig{
			"service1": {},
		},
	})
	return &inspector{
		dockerClient: client,
		executor:     mockExec,
		dispatcher:   mockDispatcher,
		servicesDir:  servicesDir,
		configStore:  configStore,
	}
}

func TestGetManagedStacks(t *testing.T) {
	mockClient := new(MockClient)
	mockExec := new(MockExec)
	mockDispatcher := new(MockDispatcher)

	// Test successful case
	mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{
		Items: []container.Summary{
			{
				ID:     "container1",
				Names:  []string{"/container1"},
				Image:  "image1",
				State:  "running",
				Status: "Up 1 hour",
			},
			{
				ID:     "container2",
				Names:  []string{"/container2"},
				Image:  "image2",
				State:  "exited",
				Status: "Exited (0) 2 hours ago",
			},
		},
	}, nil)

	mockClient.On("ContainerInspect", mock.Anything, "container1", mock.Anything).Return(client.ContainerInspectResult{
		Container: container.InspectResponse{
			Config: &container.Config{
				Labels: map[string]string{
					"com.docker.compose.project.working_dir": "/services/service1",
				},
			},
			State: &container.State{
				Health: &container.Health{
					Status: container.Healthy,
				},
				StartedAt: "2006-01-02T15:04:05.999999999Z",
			},
		},
	}, nil)

	mockClient.On("ContainerInspect", mock.Anything, "container2", mock.Anything).Return(client.ContainerInspectResult{
		Container: container.InspectResponse{
			Config: &container.Config{
				Labels: map[string]string{
					"com.docker.compose.project.working_dir": "/services/service2",
				},
			},
			State: &container.State{
				Health: &container.Health{
					Status: container.Healthy,
				},
			},
		},
	}, errors.New("failed to inspect container"))

	servicesDir := "/services"
	inspector := newInspectorWithMock(t, mockClient, mockExec, mockDispatcher, servicesDir)
	result, err := inspector.GetManagedStacks()

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, "service1")
	assert.Len(t, result["service1"], 1)

	// Test error case
	mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{}, errors.New("failed to list containers"))

	_, err = inspector.GetManagedStacks()

	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to list containers")
}

func TestGetServiceNameFromLabel(t *testing.T) {
	testCases := []struct {
		name           string
		labels         map[string]string
		expectedResult string
	}{
		{
			name: "Successful case",
			labels: map[string]string{
				"com.docker.compose.project.working_dir": "/services/service1",
			},
			expectedResult: "service1",
		},

		{
			name:           "Label not found",
			labels:         map[string]string{},
			expectedResult: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inspectResponse := client.ContainerInspectResult{
				Container: container.InspectResponse{
					Config: &container.Config{
						Labels: tc.labels,
					},
				},
			}
			servicesDir := "/services"
			serviceName := getServiceNameFromLabel(inspectResponse, servicesDir)

			assert.Equal(t, tc.expectedResult, serviceName)
		})
	}
}

func TestGetServiceContainers(t *testing.T) {
	mockClient := new(MockClient)
	mockExec := new(MockExec)
	mockDispatcher := new(MockDispatcher)

	// Test successful case
	mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service1", "config", "--services"}).Once().Return([]byte("service1 service2"), nil)

	servicesDir := "/services"
	serviceName := "service1"
	inspector := newInspectorWithMock(t, mockClient, mockExec, mockDispatcher, servicesDir)
	result, err := inspector.getServiceContainers(serviceName)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, "service1")
	assert.Contains(t, result, "service2")

	// Test error case
	mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service2", "config", "--services"}).Once().Return([]byte{}, errors.New("failed to get services"))

	_, err = inspector.getServiceContainers("service2")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to get services")
}

func TestGetCurrentStacksState(t *testing.T) {
	t.Run("successful case with healthy stack", func(t *testing.T) {
		mockClient := new(MockClient)
		mockExec := new(MockExec)
		mockDispatcher := new(MockDispatcher)

		servicesDir := "/services"
		inspector := newInspectorWithMock(t, mockClient, mockExec, mockDispatcher, servicesDir)

		mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{
			Items: []container.Summary{
				{
					ID:     "container1",
					Names:  []string{"/container1"},
					Image:  "image1",
					State:  "running",
					Status: "Up 1 hour",
					Labels: map[string]string{
						"com.docker.compose.service": "service1",
					},
				},
			},
		}, nil)

		mockClient.On("ContainerInspect", mock.Anything, "container1", mock.Anything).Return(client.ContainerInspectResult{
			Container: container.InspectResponse{
				Config: &container.Config{
					Labels: map[string]string{
						"com.docker.compose.project.working_dir": "/services/service1",
						"com.docker.compose.service":             "service1",
					},
				},
				State: &container.State{
					Health: &container.Health{
						Status: container.Healthy,
					},
					StartedAt: "2006-01-02T15:04:05.999999999Z",
				},
			},
		}, nil)

		mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service1", "config", "--services"}).Once().Return([]byte("service1"), nil)

		cfg := models.Config{
			Services: map[string]models.ServiceConfig{
				"service1": {
					"Enabled": "true",
				},
			},
		}

		result, err := inspector.GetCurrentStacksState(cfg.GetEnabledServices())

		assert.NoError(t, err)
		assert.Equal(t, models.StackStatusHealthy, result.ForService("service1"))
	})

	t.Run("case with unhealthy stack (missing container)", func(t *testing.T) {
		mockClient := new(MockClient)
		mockExec := new(MockExec)
		mockDispatcher := new(MockDispatcher)
		servicesDir := "/services"
		inspector := newInspectorWithMock(t, mockClient, mockExec, mockDispatcher, servicesDir)

		mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{
			Items: []container.Summary{},
		}, nil)

		mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service2", "config", "--services"}).Once().Return([]byte("service2"), nil)

		cfg := models.Config{
			Services: map[string]models.ServiceConfig{
				"service2": {
					"Enabled": "true",
				},
			},
		}

		result, err := inspector.GetCurrentStacksState(cfg.GetEnabledServices())

		assert.NoError(t, err)
		assert.Equal(t, models.StackStatusUnhealthy, result.ForService("service2"))
	})

	t.Run("case with error getting service containers", func(t *testing.T) {
		mockClient := new(MockClient)
		mockExec := new(MockExec)
		mockDispatcher := new(MockDispatcher)
		servicesDir := "/services"
		inspector := newInspectorWithMock(t, mockClient, mockExec, mockDispatcher, servicesDir)
		mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{
			Items: []container.Summary{},
		}, nil)
		mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service3", "config", "--services"}).Once().Return([]byte{}, errors.New("failed to get services"))

		cfg := models.Config{
			Services: map[string]models.ServiceConfig{
				"service3": {
					"Enabled": "true",
				},
			},
		}

		result, err := inspector.GetCurrentStacksState(cfg.GetEnabledServices())

		assert.NoError(t, err)
		assert.Equal(t, models.StackStatusUnhealthy, result.ForService("service3"))
	})

}

func TestStateRefresh_TriggersUnhealthyEvent(t *testing.T) {

	mockClient := new(MockClient)
	mockExec := new(MockExec)
	mockDispatcher := new(MockDispatcher)
	servicesDir := "/services"
	inspector := newInspectorWithMock(t, mockClient, mockExec, mockDispatcher, servicesDir)
	inspector._refreshDuration = time.Millisecond

	mockClient.On("ContainerList", mock.Anything, mock.Anything).Return(client.ContainerListResult{
		Items: []container.Summary{
			{
				ID:     "service1",
				Names:  []string{"/service1"},
				Image:  "image1",
				State:  "exited",
				Status: "Up 1 hour",
				Labels: map[string]string{
					"com.docker.compose.service": "service1",
				},
			},
		},
	}, nil)
	mockClient.On("ContainerInspect", mock.Anything, "service1", mock.Anything).Return(client.ContainerInspectResult{
		Container: container.InspectResponse{
			Config: &container.Config{
				Labels: map[string]string{
					"com.docker.compose.project.working_dir": "/services/service1",
					"com.docker.compose.service":             "service1",
				},
			},
			State: &container.State{
				Health: &container.Health{
					Status: container.Unhealthy,
				},
				StartedAt: "2006-01-02T15:04:05.999999999Z",
			},
		},
	}, nil)
	mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service1", "config", "--services"}).Return([]byte("service1"), nil)

	// Expect dispatcher to be called with STACKS_UNHEALTHY
	mockDispatcher.On("Dispatch", mock.Anything, models.EventStacksUnhealthy, mock.Anything).Once()

	// Run scheduleStateRefresh in a goroutine
	inspector.refreshState()

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, models.StackStatusUnhealthy, inspector.currentStacksState.GlobalStatus)

	mockDispatcher.AssertExpectations(t)
}

func TestStateRefresh_DoesNotTriggerEventWhenHealthy(t *testing.T) {
	mockClient := new(MockClient)
	mockExec := new(MockExec)
	mockDispatcher := new(MockDispatcher)
	servicesDir := "/services"
	inspector := newInspectorWithMock(t, mockClient, mockExec, mockDispatcher, servicesDir)
	inspector._refreshDuration = time.Millisecond

	mockClient.On("ContainerList", mock.Anything, mock.Anything).Return(client.ContainerListResult{
		Items: []container.Summary{
			{
				ID:     "service1",
				Names:  []string{"/service1"},
				Image:  "image1",
				State:  "running",
				Status: "Up 1 hour",
				Labels: map[string]string{
					"com.docker.compose.service": "service1",
				},
			},
		},
	}, nil)
	mockClient.On("ContainerInspect", mock.Anything, "service1", mock.Anything).Return(client.ContainerInspectResult{
		Container: container.InspectResponse{
			Config: &container.Config{
				Labels: map[string]string{
					"com.docker.compose.project.working_dir": "/services/service1",
					"com.docker.compose.service":             "service1",
				},
			},
			State: &container.State{
				Health: &container.Health{
					Status: container.Healthy,
				},
				StartedAt: "2006-01-02T15:04:05.999999999Z",
			},
		},
	}, nil)
	mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service1", "config", "--services"}).Return([]byte("service1"), nil)
	inspector.refreshState()

	time.Sleep(10 * time.Millisecond)

	mockDispatcher.AssertNotCalled(t, "Dispatch", mock.Anything, models.EventStacksUnhealthy, mock.Anything)
	assert.Equal(t, models.StackStatusHealthy, inspector.currentStacksState.GlobalStatus)
}
