package docker

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"

	"github.com/moby/moby/client"
)

// Inspector defined operations for info retreival on containers
type Inspector interface {
	GetManagedStacks() (map[string][]models.ContainerSummary, error)
	GetStacksState() (models.StacksState, error)
	GetCurrentStacksState(services []string) (models.StacksState, error)
}

// Client defines the methods from the Docker client that are used by the Inspector
type Client interface {
	ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error)
	ContainerInspect(ctx context.Context, containerID string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error)
}

// inspector implements information retrieval about docker stacks
type inspector struct {
	configStore  storage.ConfigStore
	dispatcher   events.Dispatcher
	executor     shell.Executor
	dockerClient Client
	servicesDir  string

	currentStacksState models.StacksState
	_refreshDuration   time.Duration
}

// NewInspector creates new inspector given a docker client
func NewInspector(servicesDir string, configStore storage.ConfigStore, dispatcher events.Dispatcher) (Inspector, error) {
	client, err := client.New(client.FromEnv)
	if err != nil {
		slog.Error("Failed to create docker client", "error", err)
		return nil, err
	}
	inspector := inspector{
		configStore:      configStore,
		dispatcher:       dispatcher,
		executor:         shell.NewExecutor(),
		dockerClient:     client,
		servicesDir:      servicesDir,
		_refreshDuration: 3 * time.Minute,
	}
	go inspector.scheduleStateRefresh() // should be move somewhere else, and should provide stopping mechanism
	return &inspector, nil
}

// GetManagedStacks returns the list of containers (as returned by ContainerList)
// that are managed by AirCompose
func (i *inspector) GetManagedStacks() (map[string][]models.ContainerSummary, error) {
	cfg, err := i.configStore.Get()
	if err != nil {
		return nil, err
	}
	return i.getRunningStacks(cfg.GetEnabledServices())
}

func (i *inspector) getRunningStacks(services []string) (map[string][]models.ContainerSummary, error) {
	ctx := context.Background()
	summaries, err := i.dockerClient.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	matches := make(map[string][]models.ContainerSummary)
	for _, service := range services {
		matches[service] = []models.ContainerSummary{}
	}
	for _, c := range summaries.Items {

		inspect, err := i.dockerClient.ContainerInspect(ctx, c.ID, client.ContainerInspectOptions{})
		if err != nil {
			slog.Error("Failed to inspect container",
				"containerId", c.ID, "names", c.Names, "error", err)
			continue
		}
		serviceName := getServiceNameFromLabel(inspect, i.servicesDir)
		if !slices.Contains(services, serviceName) {
			continue
		}
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

func (i *inspector) getServiceContainers(serviceName string) ([]string, error) {
	result, err := i.executor.Exec("docker", "compose", "--project-directory", filepath.Join(i.servicesDir, serviceName), "config", "--services")
	return strings.Fields(string(result)), err
}

func (i *inspector) GetStacksState() (models.StacksState, error) {
	return i.currentStacksState, nil
}

func (i *inspector) scheduleStateRefresh() {
	ticker := time.NewTicker(i._refreshDuration)
	defer ticker.Stop()

	for range ticker.C {
		i.refreshState()
	}
}

func (i *inspector) refreshState() {
	cfg, err := i.configStore.Get()
	if err != nil {
		slog.Error("error while getting configuration in health refresh", "err", err)
		return
	}
	state, err := i.GetCurrentStacksState(cfg.GetEnabledServices())
	if err != nil {
		slog.Error("error while getting stacks state", "err", err)
		return
	}
	i.setCurrentState(state)
}

func (i *inspector) setCurrentState(newState models.StacksState) {
	if i.currentStacksState.GlobalStatus != newState.GlobalStatus {
		if newState.GlobalStatus == models.StackStatusUnhealthy {
			i.dispatcher.Dispatch(context.Background(), models.EventStacksUnhealthy,

				fmt.Sprintf("containers : %v", strings.Join(newState.GetUnhealthyContainers(), ", ")))
		}
	}
	i.currentStacksState = newState
}

func (i *inspector) GetCurrentStacksState(services []string) (models.StacksState, error) {
	state := models.NewStacksState()
	runningStacks, err := i.getRunningStacks(services)
	if err != nil {
		return models.StacksState{}, err
	}

	for service, serviceContainers := range runningStacks {

		expectedContainers, err := i.getServiceContainers(service)
		slog.Debug("expectedServices ", "service", service, "expectedServices", expectedContainers, "err", err)
		if err != nil {
			state.ProgressiveUpdateServiceStatus(service, models.StackStatusUnhealthy)
			continue
		}
		slog.Debug("running containers ", "service", service, "serviceContainers", serviceContainers)

		if len(serviceContainers) != len(expectedContainers) {
			state.ProgressiveUpdateServiceStatus(service, models.StackStatusUnhealthy)
			continue
		}
		for _, ctr := range serviceContainers {
			if !slices.Contains(expectedContainers, ctr.Name) {
				state.ProgressiveUpdateServiceStatus(service, models.StackStatusUnhealthy)
				break
			}
			state.CombineContainerStatus(service, ctr)
		}
	}
	slog.Debug(fmt.Sprintf("all services are %+v", state))
	return state, nil
}
