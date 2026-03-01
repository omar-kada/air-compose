package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"omar-kada/air-compose/internal/shell"

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

func newInspectorWithMock(client Client, mockExec shell.Executor) *inspector {
	return &inspector{
		log:          slog.Default(),
		dockerClient: client,
		executor:     mockExec,
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
	inspector := newInspectorWithMock(mockClient, mockExec)
	result, err := inspector.GetManagedStacks(servicesDir)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, "service1")
	assert.Len(t, result["service1"], 1)

	// Test error case
	mockClient.On("ContainerList", mock.Anything, mock.Anything).Once().Return(client.ContainerListResult{}, errors.New("failed to list containers"))

	_, err = inspector.GetManagedStacks(servicesDir)

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

	// Test successful case
	mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service1", "config", "--services"}).Once().Return([]byte("service1 service2"), nil)

	servicesDir := "/services"
	serviceName := "service1"
	inspector := newInspectorWithMock(mockClient, mockExec)
	result, err := inspector.GetServiceContainers(serviceName, servicesDir)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, "service1")
	assert.Contains(t, result, "service2")

	// Test error case
	mockExec.On("Exec", "docker", []string{"compose", "--project-directory", "/services/service2", "config", "--services"}).Once().Return([]byte{}, errors.New("failed to get services"))

	_, err = inspector.GetServiceContainers("service2", servicesDir)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to get services")
}
