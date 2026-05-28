package docker

import (
	"context"
	"errors"
	"testing"

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

type MockExec struct {
	mock.Mock
}

func (m *MockExec) Exec(cmd string, cmdArgs ...string) ([]byte, error) {
	args := m.Called(cmd, cmdArgs)
	return args.Get(0).([]byte), args.Error(1)
}

func newInspectorWithMock(t *testing.T, client Client, mockExec shell.Executor, servicesDir string) *inspector {
	configStore := storage.NewConfigStore(t.TempDir() + "/config.yaml")
	configStore.Update(models.Config{
		Services: map[string]models.ServiceConfig{
			"service1": {},
		},
	})
	return &inspector{
		dockerClient: client,
		executor:     mockExec,
		servicesDir:  servicesDir,
		configStore:  configStore,
	}
}

func TestGetManagedStacks(t *testing.T) {
	mockClient := new(MockClient)
	mockExec := new(MockExec)

	// Test successful case
	mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{
		Items: []container.Summary{
			{
				ID:     "container1",
				Names:  []string{"container1"},
				Image:  "image1",
				State:  "running",
				Status: "Up 1 hour (healthy)",
				Labels: map[string]string{
					"com.docker.compose.project.working_dir": "/services/service1",
					"com.docker.compose.service":             "container1",
				},
			},
			{
				ID:     "container2",
				Names:  []string{"container2"},
				Image:  "image2",
				State:  "exited",
				Status: "Exited (0) 2 hours ago",
				Labels: map[string]string{
					"com.docker.compose.project.working_dir": "/services/service1",
					"com.docker.compose.service":             "container2",
				},
			},
		},
	}, nil)

	mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service1", "config", "--services"}).
		Return([]byte("container1 container2"), nil)

	servicesDir := "/services"
	inspector := newInspectorWithMock(t, mockClient, mockExec, servicesDir)
	result, err := inspector.GetManagedStacks()

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, "service1")
	assert.Len(t, result["service1"], 2)
	// Add assertions for each container status
	assert.Equal(t, models.ContainerHealthy, result["service1"]["container1"].Health)
	assert.Equal(t, models.ContainerUnhealthy, result["service1"]["container2"].Health)

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
			container := container.Summary{
				Labels: tc.labels,
			}
			servicesDir := "/services"
			serviceName := getServiceNameFromLabel(container, servicesDir)

			assert.Equal(t, tc.expectedResult, serviceName)
		})
	}
}

func TestGetServiceContainers(t *testing.T) {
	mockClient := new(MockClient)
	mockExec := new(MockExec)

	// Test successful case
	mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service1", "config", "--services"}).Once().Return([]byte("service1 service2"), nil)

	servicesDir := "/services"
	serviceName := "service1"
	inspector := newInspectorWithMock(t, mockClient, mockExec, servicesDir)
	result, err := inspector.getExpectedServiceContainers(serviceName)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, "service1")
	assert.Contains(t, result, "service2")

	// Test error case
	mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service2", "config", "--services"}).Once().Return([]byte{}, errors.New("failed to get services"))

	_, err = inspector.getExpectedServiceContainers("service2")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to get services")
}

func TestGetCurrentStacksState(t *testing.T) {
	t.Run("successful case with healthy stack", func(t *testing.T) {
		mockClient := new(MockClient)
		mockExec := new(MockExec)

		servicesDir := "/services"
		inspector := newInspectorWithMock(t, mockClient, mockExec, servicesDir)

		mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{
			Items: []container.Summary{
				{
					ID:     "container1",
					Names:  []string{"container1"},
					Image:  "image1",
					State:  container.StateRunning,
					Status: "Up 1 hour (healthy)",
					Labels: map[string]string{
						"com.docker.compose.service":             "container1",
						"com.docker.compose.project.working_dir": "/services/service1",
					},
				},
			},
		}, nil)

		mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service1", "config", "--services"}).Once().Return([]byte("container1"), nil)

		cfg := models.Config{
			Services: map[string]models.ServiceConfig{
				"service1": {
					"Enabled": "true",
				},
			},
		}

		result, err := inspector.GetCurrentStacks(cfg.GetEnabledServices())

		assert.NoError(t, err)
		assert.Equal(t, models.ContainerHealthy, result.GetGlobalHealth())
		containerStatue := result["service1"]["container1"]
		assert.Equal(t, models.ContainerHealthy, containerStatue.Health)
	})

	t.Run("case with unhealthy stack (missing container)", func(t *testing.T) {
		mockClient := new(MockClient)
		mockExec := new(MockExec)
		servicesDir := "/services"
		inspector := newInspectorWithMock(t, mockClient, mockExec, servicesDir)

		mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{
			Items: []container.Summary{},
		}, nil)

		mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service2", "config", "--services"}).Once().Return([]byte("container2"), nil)

		cfg := models.Config{
			Services: map[string]models.ServiceConfig{
				"service2": {
					"Enabled": "true",
				},
			},
		}

		result, err := inspector.GetCurrentStacks(cfg.GetEnabledServices())

		assert.NoError(t, err)
		assert.Equal(t, models.ContainerUnhealthy, result.GetGlobalHealth())
		assert.NotContains(t, result["service2"], "container1")
	})

	t.Run("case with error getting service containers", func(t *testing.T) {
		mockClient := new(MockClient)
		mockExec := new(MockExec)
		servicesDir := "/services"
		inspector := newInspectorWithMock(t, mockClient, mockExec, servicesDir)
		mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{
			Items: []container.Summary{},
		}, nil)
		mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service3", "config", "--services"}).
			Once().
			Return([]byte{}, errors.New("failed to get services"))

		cfg := models.Config{
			Services: map[string]models.ServiceConfig{
				"service3": {
					"Enabled": "true",
				},
			},
		}

		result, err := inspector.GetCurrentStacks(cfg.GetEnabledServices())

		assert.NoError(t, err)
		assert.Equal(t, models.ContainerUnhealthy, result.GetGlobalHealth())
		assert.NotContains(t, result["service3"], "container1")
	})

}
