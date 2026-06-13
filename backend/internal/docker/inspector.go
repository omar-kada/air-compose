package docker

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/shell"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// Inspector defined operations for info retreival on containers
type Inspector interface {
	GetManagedStacks() (models.StacksState, error)
	GetCurrentStacks(services []string) (models.StacksState, error)
}

// Client defines the methods from the Docker client that are used by the Inspector
type Client interface {
	ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error)
}

// inspector implements information retrieval about docker stacks
type inspector struct {
	configStore  models.ConfigGetter
	executor     shell.Executor
	dockerClient Client
	servicesDir  string
}

// NewInspector creates new inspector given a docker client
func NewInspector(servicesDir string, configStore config.Store) (Inspector, error) {
	client, err := client.New(client.FromEnv)
	if err != nil {
		slog.Error("Failed to create docker client", "error", err)
		return nil, err
	}
	inspector := inspector{
		configStore:  configStore,
		executor:     shell.NewExecutor(),
		dockerClient: client,
		servicesDir:  servicesDir,
	}
	return &inspector, nil
}

// GetManagedStacks returns the list of containers (as returned by ContainerList)
// that are managed by AirCompose
func (i *inspector) GetManagedStacks() (models.StacksState, error) {
	return i.GetCurrentStacks(i.configStore.Get().GetEnabledServices())
}

func (i *inspector) GetCurrentStacks(services []string) (models.StacksState, error) {
	ctx := context.Background()
	summaries, err := i.dockerClient.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	matches := make(models.StacksState)
	for _, service := range services {
		matches[service] = map[string]models.ContainerSummary{}
	}
	for _, c := range summaries.Items {
		serviceName := getServiceNameFromLabel(c, i.servicesDir)
		if !slices.Contains(services, serviceName) {
			continue
		}
		if serviceName != "" {
			ctrName := c.Labels["com.docker.compose.service"]
			matches.SetContainerStatus(serviceName, models.ContainerSummary{
				ID:        c.ID,
				Name:      ctrName,
				Image:     c.Image,
				State:     models.ContainerState(c.State),
				Health:    parseHealthStatus(c.Status, models.ContainerState(c.State)),
				StartedAt: time.Unix(c.Created, 0),
			})
		}
	}
	for _, service := range services {

		expectedContainers, err := i.getExpectedServiceContainers(service)
		if err != nil {
			slog.Error("error getting expected containers ", "service", service, "expectedContainers", expectedContainers, "err", err)
			matches.SetContainerStatus(service, models.ContainerSummary{
				Name:   "",
				State:  models.StateDead,
				Health: models.ContainerUnhealthy,
			})
			continue
		}

		for _, ctrName := range expectedContainers {
			_, found := matches[service][ctrName]
			if !found {
				matches.SetContainerStatus(service, models.ContainerSummary{
					Name:   ctrName,
					State:  models.StateDead,
					Health: models.ContainerUnhealthy,
				})
			}
		}
	}
	return matches, nil
}

func getServiceNameFromLabel(container container.Summary, servicesDir string) string {
	if value, ok := container.Labels["com.docker.compose.project.working_dir"]; ok {
		if after, found := strings.CutPrefix(value, servicesDir); found {
			return strings.TrimPrefix(after, "/")
		}
	}
	return ""
}

func parseHealthStatus(status string, state models.ContainerState) models.ContainerHealth {
	switch {
	case strings.Contains(status, "(healthy)"):
		return models.ContainerHealthy
	case strings.Contains(status, "(unhealthy)"):
		return models.ContainerUnhealthy
	case strings.Contains(status, "(health: starting)"):
		return models.ContainerStarting
	default:
		switch state {
		case models.StateDead, models.StateExited, models.StateRemoving, models.StatePaused:
			return models.ContainerUnhealthy
		default:
			return models.ContainerNoHealth
		}
	}
}

func (i *inspector) getExpectedServiceContainers(serviceName string) ([]string, error) {
	result, err := i.executor.Exec("docker", "compose", "--project-directory", filepath.Join(i.servicesDir, serviceName), "config", "--services")
	return strings.Fields(string(result)), err
}
