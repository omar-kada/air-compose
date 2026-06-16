package docker

import (
	"errors"
	"testing"

	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/testutil"
	"omar-kada/air-compose/testutil/mocks"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newInspectorWithMock(t *testing.T, client Client, mockExec shell.Executor, servicesDir string) *inspector {
	t.Helper()
	configStore := testutil.NewConfigGetter(models.Config{
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
	mockClient := new(mocks.DockerClient)
	mockExec := new(mocks.Executor)

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

	mockExec.On("NoLogs").Return(mockExec)

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
	mockClient := new(mocks.DockerClient)
	mockExec := new(mocks.Executor)

	// Test successful case
	mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service1", "config", "--services"}).Once().Return([]byte("service1 service2"), nil)
	mockExec.On("NoLogs").Return(mockExec)

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
		mockClient := new(mocks.DockerClient)
		mockExec := new(mocks.Executor)

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
		mockExec.On("NoLogs").Return(mockExec)

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
		mockClient := new(mocks.DockerClient)
		mockExec := new(mocks.Executor)
		servicesDir := "/services"
		inspector := newInspectorWithMock(t, mockClient, mockExec, servicesDir)

		mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{
			Items: []container.Summary{},
		}, nil)

		mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service2", "config", "--services"}).Once().Return([]byte("container2"), nil)
		mockExec.On("NoLogs").Return(mockExec)

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
		mockClient := new(mocks.DockerClient)
		mockExec := new(mocks.Executor)
		servicesDir := "/services"
		inspector := newInspectorWithMock(t, mockClient, mockExec, servicesDir)
		mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{
			Items: []container.Summary{},
		}, nil)
		mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service3", "config", "--services"}).
			Once().
			Return([]byte{}, errors.New("failed to get services"))
		mockExec.On("NoLogs").Return(mockExec)

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

func TestParseHealthStatus(t *testing.T) {
	testCases := []struct {
		name           string
		status         string
		state          models.ContainerState
		expectedHealth models.ContainerHealth
	}{
		{
			name:           "healthy container",
			status:         "Up 1 hour (healthy)",
			state:          models.StateRunning,
			expectedHealth: models.ContainerHealthy,
		},
		{
			name:           "unhealthy container",
			status:         "Up 1 hour (unhealthy)",
			state:          models.StateRunning,
			expectedHealth: models.ContainerUnhealthy,
		},
		{
			name:           "starting container",
			status:         "Up 1 hour (health: starting)",
			state:          models.StateRunning,
			expectedHealth: models.ContainerStarting,
		},
		{
			name:           "dead container",
			status:         "Exited (0) 2 hours ago",
			state:          models.StateDead,
			expectedHealth: models.ContainerUnhealthy,
		},
		{
			name:           "exited container",
			status:         "Exited (0) 2 hours ago",
			state:          models.StateExited,
			expectedHealth: models.ContainerUnhealthy,
		},
		{
			name:           "removing container",
			status:         "Removing",
			state:          models.StateRemoving,
			expectedHealth: models.ContainerUnhealthy,
		},
		{
			name:           "paused container",
			status:         "Paused",
			state:          models.StatePaused,
			expectedHealth: models.ContainerUnhealthy,
		},
		{
			name:           "running container with no health status",
			status:         "Up 1 hour",
			state:          models.StateRunning,
			expectedHealth: models.ContainerNoHealth,
		},
		{
			name:           "created container",
			status:         "Created",
			state:          models.StateCreated,
			expectedHealth: models.ContainerNoHealth,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			health := parseHealthStatus(tc.status, tc.state)
			assert.Equal(t, tc.expectedHealth, health)
		})
	}
}
