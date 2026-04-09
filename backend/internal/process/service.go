// Package process handles the deployment and management of services.
package process

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"sync"

	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"
)

const (
	// WorkingBranch is the branch used for temporary deployment changes
	WorkingBranch = "to_be_deployed"
)

// Service abstracts service deployment operations
type Service interface {
	SyncDeployment() (models.Deployment, error)
	GetCurrentState() (models.State, error)
	GetDiff() ([]models.FileDiff, error)
	GetManagedStacks() (map[string][]models.ContainerSummary, error)
	GetDeployments(limit int, offset uint64) ([]models.Deployment, error)
	GetDeployment(id uint64) (models.Deployment, error)
	GetNotifications(limit int, offset uint64) ([]models.Event, error)
}

// NewService creates a new process Service instance
func NewService(
	deployParams models.DeploymentParams,
	containersDeployer docker.Deployer,
	containersInspector docker.Inspector,
	fetcher git.Fetcher,
	store storage.DeploymentStorage,
	eventStore storage.EventStorage,
	configStore storage.ConfigStore,
	dispatcher events.Dispatcher,
	scheduler ConfigScheduler,
) Service {
	cfg, _ := configStore.Get()
	return &service{
		containersDeployer:  containersDeployer,
		containersInspector: containersInspector,
		fetcher:             fetcher,
		store:               store,
		eventStore:          eventStore,
		configStore:         configStore,
		dispatcher:          dispatcher,
		params:              deployParams,
		scheduler:           scheduler,
		currentCfg:          cfg,
	}
}

// service is responsible for deploying the services
type service struct {
	containersDeployer  docker.Deployer
	containersInspector docker.Inspector
	fetcher             git.Fetcher
	store               storage.DeploymentStorage
	eventStore          storage.EventStorage
	configStore         storage.ConfigStore
	dispatcher          events.Dispatcher
	scheduler           ConfigScheduler
	params              models.DeploymentParams

	currentCfg models.Config
	mu         sync.Mutex
}

func (s *service) SyncDeployment() (models.Deployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.configStore.Get()
	if err != nil || cfg.Settings.Repo == "" {
		return models.Deployment{}, fmt.Errorf("error getting repo: %v, %w", cfg.Settings.Repo, err)
	}
	oldCfg := s.currentCfg
	s.currentCfg = cfg
	slog.Info("deploying from " + cfg.Settings.Repo + "/" + cfg.GetBranch())

	patch, syncErr := s.fetcher.DiffWithRemote()

	if syncErr != nil && syncErr != git.NoErrAlreadyUpToDate {
		return models.Deployment{}, fmt.Errorf("error getting config repo:  %w", syncErr)
	}

	// check if the config changed from last run
	configChanged := !reflect.DeepEqual(oldCfg, cfg)
	healthyStacks := s.areStacksHealthy(cfg)
	if patch.Diff == "" && !configChanged && healthyStacks {
		slog.Info("Configuration and repository are up to date. No changes detected.",
			"oldConfig", oldCfg, "newConfig", cfg, "diff", patch.Diff)
		return models.Deployment{}, nil
	}
	title := patch.Title
	if title == "" {
		if configChanged {
			title = "Configuration changed"
		} else if !healthyStacks {
			title = "Unhealthy stacks"
		} else {
			title = "Manual Deploy"
		}
	}

	deployment, err := s.store.InitDeployment(title, patch.Author, patch.Diff, patch.Files)
	ctx := events.GetDeploymentContext(context.Background(), deployment)
	s.dispatcher.Dispatch(ctx, models.EventDeploymentStarted, "")
	if err != nil {
		return deployment, err
	}
	go func() {
		err := s.fetcher.PullBranch(WorkingBranch, "")
		if err != nil {
			s.updateDeploymentStatus(ctx, deployment, err)
			return
		}
		s.dispatcher.Dispatch(ctx, models.EventMisc, "Pulled new changes into working branch")

		err = s.containersDeployer.WithCtx(ctx).RemoveAndDeployStacks(oldCfg, cfg, s.params)
		if err != nil {
			s.updateDeploymentStatus(ctx, deployment, err)
			return
		}

		err = s.fetcher.PullBranch(cfg.GetBranch(), patch.CommitHash)
		s.updateDeploymentStatus(ctx, deployment, err)
	}()

	return deployment, nil
}

func (s *service) areStacksHealthy(cfg models.Config) bool {
	state, err := s.getStacksState(cfg)
	if err != nil {
		return false
	}
	return state.GetGlobalHealth() == models.StackStatusHealthy || state.GetGlobalHealth() == models.StackStatusStarting
}

func (s *service) getStacksState(cfg models.Config) (models.StacksState, error) {
	state := models.NewStacksState()
	runningStacks, err := s.containersInspector.GetManagedStacks(s.params.ServicesDir)
	if err != nil {
		return state, err
	}
	enabledServices := cfg.GetEnabledServices()

enabledServiceLoop:
	for _, service := range enabledServices {

		expectedContainers, err := s.containersInspector.GetServiceContainers(service, s.params.ServicesDir)
		slog.Debug("expectedServices ", "service", service, "expectedServices", expectedContainers, "err", err)
		if err != nil {
			state.ProgressiveUpdateServiceStatus(service, models.StackStatusUnhealthy)
			continue
		}
		serviceContainers := runningStacks[service]
		slog.Debug("running containers ", "service", service, "serviceContainers", serviceContainers)

		if len(serviceContainers) != len(expectedContainers) {
			state.ProgressiveUpdateServiceStatus(service, models.StackStatusUnhealthy)
			continue
		}
		for _, ctr := range serviceContainers {

			if !slices.Contains(expectedContainers, ctr.Name) {
				state.ProgressiveUpdateServiceStatus(service, models.StackStatusUnhealthy)
				continue enabledServiceLoop
			}
			state.CombineContainerStatus(service, ctr)
		}
	}
	slog.Debug(fmt.Sprintf("all services are %+v", state))
	return state, nil
}

func (s *service) updateDeploymentStatus(ctx context.Context, deployment models.Deployment, err error) {
	if err != nil {
		s.dispatcher.Dispatch(ctx, models.EventDeploymentError, err.Error())
		s.store.EndDeployment(deployment.ID, models.DeploymentStatusError)
	} else {
		s.dispatcher.Dispatch(ctx, models.EventDeploymentSuccess, "")
		s.store.EndDeployment(deployment.ID, models.DeploymentStatusSuccess)
	}
}

// GetManagedStacks returns a map of all containers managed by the tool
func (s *service) GetManagedStacks() (map[string][]models.ContainerSummary, error) {
	return s.containersInspector.GetManagedStacks(s.params.ServicesDir)
}

// GetCurrentState returns the statistics of deployments for the last N days
func (s *service) GetCurrentState() (models.State, error) {
	dep, _ := s.store.GetLastDeployment()
	stackstate, _ := s.getStacksState(s.currentCfg)
	cfg, _ := s.configStore.Get()

	return models.State{
		LastStatus:  dep.Status,
		NextDeploy:  s.scheduler.GetNext(),
		Health:      stackstate.GetGlobalHealth(),
		Initialized: cfg.Settings.Repo != "",
	}, nil
}

// GetDiff returns the changed files between what's deployed and the repo
func (s *service) GetDiff() ([]models.FileDiff, error) {
	patch, err := s.fetcher.DiffWithRemote()
	if err != nil {
		return []models.FileDiff{}, err
	}
	return patch.Files, nil
}

// GetDeployments returns a paginated list of deployments.
func (s *service) GetDeployments(limit int, offset uint64) ([]models.Deployment, error) {
	return s.store.GetDeployments(storage.NewIDCursor(limit, offset))
}

// GetNotifications returns a paginated list of notifications.
func (s *service) GetNotifications(limit int, offset uint64) ([]models.Event, error) {
	return s.eventStore.GetNotifications(storage.NewIDCursor(limit, offset))
}

// GetDeployment returns a deployment.
func (s *service) GetDeployment(id uint64) (models.Deployment, error) {
	return s.store.GetDeployment(id)
}
