package mocks

import (
	"context"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/mock"
)

// DockerClient is a mock implementation of the Client interface
type DockerClient struct {
	mock.Mock
}

// ContainerList mocks the ContainerList method of the Docker client.
func (m *DockerClient) ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(client.ContainerListResult), args.Error(1)
}
