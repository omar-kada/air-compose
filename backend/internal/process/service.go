// Package process handles the deployment and management of services.
package process

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
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
}

// NewService creates a new process Service instance
func NewService(
	deployParams models.DeploymentParams,
	containersDeployer docker.Deployer,
	containersInspector docker.Inspector,
	fetcher git.Fetcher,
	store storage.DeploymentStorage,
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
	if err != nil || cfg.Settings.Git.Repo == "" {
		return models.Deployment{}, fmt.Errorf("error getting repo: %v, %w", cfg.Settings.Git.Repo, err)
	}
	oldCfg := s.currentCfg
	s.currentCfg = cfg
	slog.Info("deploying from " + cfg.Settings.Git.Repo + "/" + cfg.GetBranch())

	patch, syncErr := s.fetcher.DiffWithRemote()

	if syncErr != nil && syncErr != git.NoErrAlreadyUpToDate {
		return models.Deployment{}, fmt.Errorf("error getting config repo:  %w", syncErr)
	}

	// check if the config changed from last run
	configChanged := !reflect.DeepEqual(oldCfg, cfg)
	healthyStacks := s.areStacksHealthy(cfg)
	if patch.Diff == "" && !configChanged && healthyStacks {
		slog.Debug("Configuration and repository are up to date. No changes detected.")
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

	deployment, err := s.store.InitDeployment(title, patch, cfg.Settings.Git)
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
	state, err := s.containersInspector.GetCurrentStacks(cfg.GetEnabledServices())
	if err != nil {
		return false
	}
	return state.GetGlobalHealth() == models.ContainerHealthy || state.GetGlobalHealth() == models.ContainerStarting
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

// GetCurrentState returns the statistics of deployments for the last N days
func (s *service) GetCurrentState() (models.State, error) {
	dep, _ := s.store.GetLastDeployment()
	stackstate, _ := s.containersInspector.GetCurrentStacks(s.currentCfg.GetEnabledServices())
	cfg, _ := s.configStore.Get()

	return models.State{
		LastStatus:  dep.Status,
		NextDeploy:  s.scheduler.GetNext(),
		Health:      stackstate.GetGlobalHealth(),
		Initialized: cfg.Settings.Git.Repo != "",
	}, nil
}
