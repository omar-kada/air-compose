package docker

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/models"

	"github.com/moby/moby/client"
)

// Inspector defined operations for info retreival on containers
type Inspector interface {
	GetManagedStacks(servicesDir string) (map[string][]models.ContainerSummary, error)
	GetServiceContainers(serviceName string, servicesDir string) ([]string, error)
}

// Client defines the methods from the Docker client that are used by the Inspector
type Client interface {
	ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error)
	ContainerInspect(ctx context.Context, containerID string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error)
}

// inspector implements information retrieval about docker stacks
type inspector struct {
	log          *slog.Logger
	executor     shell.Executor
	dockerClient Client
}

// NewInspector creates new inspector given a docker client
func NewInspector() (Inspector, error) {
	client, err := client.New(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		slog.Error("Failed to create docker client", "error", err)
		return nil, err
	}
	return &inspector{
		log:          slog.Default(),
		executor:     shell.NewExecutor(),
		dockerClient: client,
	}, nil
}

// GetManagedStacks returns the list of containers (as returned by ContainerList)
// that are managed by AirCompose
func (i *inspector) GetManagedStacks(servicesDir string) (map[string][]models.ContainerSummary, error) {
	ctx := context.Background()
	summaries, err := i.dockerClient.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	matches := make(map[string][]models.ContainerSummary)
	for _, c := range summaries.Items {

		inspect, err := i.dockerClient.ContainerInspect(ctx, c.ID, client.ContainerInspectOptions{})
		if err != nil {
			slog.Error("Failed to inspect container",
				"containerId", c.ID, "names", c.Names, "error", err)
			continue
		}
		serviceName := getServiceNameFromLabel(inspect, servicesDir)
		if serviceName != "" {
			startedAt, err := time.Parse(time.RFC3339Nano, inspect.Container.State.StartedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse : %w", err)
			}
			matches[serviceName] = append(matches[serviceName], models.ContainerSummary{
				ID:        c.ID,
				Name:      c.Labels["com.docker.compose.service"],
				Image:     c.Image,
				State:     c.State,
				Health:    inspect.Container.State.Health.Status,
				StartedAt: startedAt,
			})
		}
	}
	return matches, nil
}

func getServiceNameFromLabel(inspect client.ContainerInspectResult, servicesDir string) string {
	for key, value := range inspect.Container.Config.Labels {
		if strings.EqualFold(key, "com.docker.compose.project.working_dir") {
			if after, found := strings.CutPrefix(value, servicesDir); found {
				return strings.TrimPrefix(after, "/")
			}
			return ""
		}
	}
	return ""
}

func (i *inspector) GetServiceContainers(serviceName string, servicesDir string) ([]string, error) {
	result, err := i.executor.Exec("docker", "compose", "--project-directory", filepath.Join(servicesDir, serviceName), "config", "--services")
	return strings.Fields(string(result)), err
}
