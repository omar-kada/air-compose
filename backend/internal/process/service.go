package process

import (
	"context"
	"errors"
	"sync"

	"omar-kada/air-compose/internal/deployments"
	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/models"
)

const (
	// WorkingBranch is the branch used for temporary deployment changes
	WorkingBranch = "to_be_deployed"
)

// ErrRepoNotDefined indicates that the repository is not defined in the configuration
var ErrRepoNotDefined = errors.New("repo is not defined in configuration")

// DeploymentTrigger represents the reason for a deployment
type DeploymentTrigger string

const (
	// DeploymentTriggerManual indicates a manual deployment
	DeploymentTriggerManual DeploymentTrigger = "manual"

	// DeploymentTriggerRepoUpdated indicates a deployment triggered by repository updates
	DeploymentTriggerRepoUpdated DeploymentTrigger = "repo_updated"

	// DeploymentTriggerConfigurationUpdated indicates a deployment triggered by configuration changes
	DeploymentTriggerConfigurationUpdated DeploymentTrigger = "configuration_updated"

	// DeploymentTriggerUnhealthyStacks indicates a deployment triggered by unhealthy stacks
	DeploymentTriggerUnhealthyStacks DeploymentTrigger = "unhealthy_stacks"
)

// DeploymentService abstracts service deployment operations
type DeploymentService interface {
	DoDeploy(trigger DeploymentTrigger, patch models.Patch) (models.Deployment, error)
}

// NewDeploymentService creates a new process Service instance
func NewDeploymentService(
	deployParams models.DeploymentParams,
	containersDeployer docker.Deployer,
	fetcher git.Fetcher,
	store deployments.DeploymentStorage,
	configStore models.ConfigGetter,
	eventPublisher events.Publisher,
) DeploymentService {
	return &service{
		containersDeployer: containersDeployer,
		fetcher:            fetcher,
		store:              store,
		configStore:        configStore,
		eventPublisher:     eventPublisher,
		params:             deployParams,
		currentCfg:         configStore.Get(),
	}
}

// service is responsible for deploying the services
type service struct {
	containersDeployer docker.Deployer
	fetcher            git.Fetcher
	store              deployments.DeploymentStorage
	configStore        models.ConfigGetter
	eventPublisher     events.Publisher
	params             models.DeploymentParams

	currentCfg models.Config
	mu         sync.Mutex
}

// DoDeploy performs a deployment based on the given trigger and patch
func (s *service) DoDeploy(trigger DeploymentTrigger, patch models.Patch) (models.Deployment, error) {

	newCfg := s.configStore.Get()
	if newCfg.Settings.Git.Repo == "" {
		return models.Deployment{}, ErrRepoNotDefined
	}
	title := patch.Title
	if title == "" {
		title = getTitleFromTrigger(trigger)
	}

	deployment, err := s.store.InitDeployment(title, patch, newCfg.Settings.Git)
	if err != nil {
		s.eventPublisher.Publish(context.Background(),
			models.SourceEvent{Type: models.EventError, Msg: err.Error()})
		return deployment, err
	}
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		ctx := models.GetDeploymentContext(context.Background(), deployment)
		var err error
		defer func() {
			// always update the deployment to success or error
			s.updateDeploymentStatus(ctx, deployment, err)
		}()

		s.eventPublisher.Publish(ctx, models.SourceEvent{Type: models.EventDeploymentStarted})
		err = s.fetcher.PullBranch(WorkingBranch, "")
		if err != nil {
			return
		}
		s.eventPublisher.Publish(ctx,
			models.SourceEvent{Type: models.EventMisc, Msg: "Pulled new changes into working branch"})

		err = s.containersDeployer.WithCtx(ctx).RemoveAndDeployStacks(s.currentCfg, newCfg, s.params)
		if err != nil {
			return
		}
		s.currentCfg = newCfg

		err = s.fetcher.PullBranch(newCfg.GetBranch(), patch.CommitHash)
	}()
	return deployment, err
}

func getTitleFromTrigger(trigger DeploymentTrigger) string {
	switch trigger {
	case DeploymentTriggerManual:
		return "Manual Deploy"
	case DeploymentTriggerRepoUpdated:
		return "Repository updated"
	case DeploymentTriggerConfigurationUpdated:
		return "Configuration changed"
	case DeploymentTriggerUnhealthyStacks:
		return "Unhealthy stacks"
	default:
		return ""
	}
}

func (s *service) updateDeploymentStatus(ctx context.Context, deployment models.Deployment, err error) {
	if err != nil {
		s.eventPublisher.Publish(ctx, models.SourceEvent{Type: models.EventDeploymentError, Msg: err.Error()})
		s.store.EndDeployment(deployment.ID, models.DeploymentStatusError)
	} else {
		s.eventPublisher.Publish(ctx, models.SourceEvent{Type: models.EventDeploymentSuccess})
		s.store.EndDeployment(deployment.ID, models.DeploymentStatusSuccess)
	}
}
