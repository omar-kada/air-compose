package testutil

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
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

// WaitForContainer waits for a container with the specified name to be ready within the given timeout.
// It returns the container summary and a boolean indicating whether the container was found.
func WaitForContainer(ctx context.Context, containerName string, timeout time.Duration) (*container.Summary, bool) {
	container, ok := WaitFor(timeout, func() (*container.Summary, bool) {
		containers, err := findContainer(ctx, containerName)
		return containers, err == nil
	})
	return container, ok
}

// WaitForDeadContainer waits for a container with the specified name to be stopped within the given timeout.
// It returns a boolean indicating whether the container was found and stopped.
func WaitForDeadContainer(ctx context.Context, containerName string, timeout time.Duration) bool {
	_, ok := WaitFor(timeout, func() (*container.Summary, bool) {
		containers, err := findContainer(ctx, containerName)
		return containers, err != nil
	})
	return ok
}

func findContainer(ctx context.Context, containerName string) (*container.Summary, error) {
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
	// Look for Git server containers
	for _, container := range containers.Items {
		//fmt.Printf("containers : %v", container.Names)
		if slices.Contains(container.Names, containerName) {
			return &container, nil
		}
	}

	return nil, errors.New("container not found " + containerName)
}

// FollowContainerLogs streams and logs container output to the test logger.
func FollowContainerLogs(ctx context.Context, t *testing.T, cli *client.Client, containerNameOrID string) {
	t.Helper()

	// Inspect once to know if the stream is multiplexed (non-TTY) or raw (TTY).
	inspectRes, err := cli.ContainerInspect(ctx, containerNameOrID, client.ContainerInspectOptions{})
	if err != nil {
		t.Logf("[%s] failed to inspect container: %v", containerNameOrID, err)
		return
	}
	isTTY := inspectRes.Container.Config != nil && inspectRes.Container.Config.Tty

	logsRes, err := cli.ContainerLogs(ctx, containerNameOrID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		t.Logf("[%s] failed to attach to container logs: %v", containerNameOrID, err)
		return
	}

	go func() {
		defer logsRes.Close()

		pr, pw := io.Pipe()

		// Demux (or passthrough) into a single pipe we can line-scan.
		go func() {
			defer pw.Close()
			var copyErr error
			if isTTY {
				_, copyErr = io.Copy(pw, logsRes)
			} else {
				_, copyErr = stdcopy.StdCopy(pw, pw, logsRes)
			}
			if copyErr != nil && !errors.Is(copyErr, io.EOF) && ctx.Err() == nil {
				t.Logf("[%s] log stream ended with error: %v", containerNameOrID, copyErr)
			}
		}()

		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // handle long lines
		for scanner.Scan() {
			line := strings.TrimRight(scanner.Text(), "\r")
			if line == "" {
				continue
			}
			t.Logf("[%s] %s", containerNameOrID, line)
		}
		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			t.Logf("[%s] log scanner error: %v", containerNameOrID, err)
		}
	}()
}

// PrintContainerLogs prints the logs for each service in the Docker Compose environment.
func PrintContainerLogs(ctx context.Context, t *testing.T, composeEnv *compose.DockerCompose) {
	t.Helper()

	for _, service := range composeEnv.Services() {

		cont, err := composeEnv.ServiceContainer(ctx, service)
		assert.NoError(t, err)

		logsReader, err := cont.Logs(ctx)
		assert.NoError(t, err)

		bytes, err := io.ReadAll(logsReader)

		assert.NoError(t, err)
		fmt.Printf("Logs for service %s:\n%s\n", service, string(bytes))
	}
}
