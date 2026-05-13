package testutil

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/testcontainers/testcontainers-go"
)

// WaitForComposeStack waits for the Docker Compose stack to be ready by checking for containers
// with the specified working directory label. It returns the list of containers and a boolean
// indicating whether the stack was found within the given timeout.
func WaitForComposeStack(ctx context.Context, workingDir string, timeout time.Duration) ([]*container.Summary, bool) {
	container, ok := WaitFor(timeout, func() ([]*container.Summary, bool) {
		containers, err := findComposeStack(ctx, workingDir)
		return containers, err == nil && len(containers) > 0
	})
	return container, ok
}

func findComposeStack(ctx context.Context, workingDir string) ([]*container.Summary, error) {
	// Create a Docker provider
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return nil, err
	}

	// List all running containers
	containers, err := provider.Client().ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	workDirLabel := "com.docker.compose.project.working_dir"
	stackContainers := make([]*container.Summary, 0)
	// Look for Git server containers
	for _, container := range containers.Items {
		if container.Labels[workDirLabel] == workingDir {
			stackContainers = append(stackContainers, &container)
		}
	}

	return stackContainers, nil
}
