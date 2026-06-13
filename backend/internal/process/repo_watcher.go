// Package process handles the deployment and management of services.
package process

import (
	"context"
	"log/slog"
	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/models"
	"time"

	"github.com/robfig/cron/v3"
)

// RepoWatcher is responsible for polling changes from repo
type RepoWatcher interface {
	Schedule() (*cron.Cron, error)
	GetNext() time.Time
}

type watcher struct {
	fetcher           git.Fetcher
	configStore       config.Store
	deploymentService DeploymentService
	eventPublisher    events.Publisher

	scheduler CronScheduler
}

// NewRepoWatcher creates a new Watcher and returns it
func NewRepoWatcher(
	fetcher git.Fetcher,
	configStore config.Store,
	deploymentService DeploymentService,
	eventPublisher events.Publisher,
	scheduler CronScheduler,
) RepoWatcher {
	return &watcher{
		fetcher:           fetcher,
		configStore:       configStore,
		deploymentService: deploymentService,
		eventPublisher:    eventPublisher,
		scheduler:         scheduler,
	}
}

func (w *watcher) DeployOnChange() {
	exists, err := w.fetcher.IsRemoteSameAsConfig()
	if err != nil {
		slog.Error("error checking repo info", "err", err)
		return
	}
	patch, syncErr := w.fetcher.DiffWithRemote()
	if exists && syncErr != nil && syncErr != git.NoErrAlreadyUpToDate {
		slog.Error("error checking new changes in repo", "err", syncErr)
		return
	}
	if exists && patch.Diff == "" {
		slog.Debug("Configuration and repository are up to date. No changes detected.")
		return
	}
	w.eventPublisher.Publish(context.Background(), models.NewNewCommitEvent(patch))
	_, err = w.deploymentService.DoDeploy(DeploymentTriggerRepoUpdated, patch)
	if err != nil {
		slog.Error("error deploying on new commit", "err", err)
	}
}

// Schedule stops the old cron when it exists, and runs a new cron job
func (w *watcher) Schedule() (*cron.Cron, error) {
	return w.scheduler.Schedule(w.DeployOnChange, w.configStore.Get().Settings.Schedule.Cron)
}

// GetNext returns the next scheduled time of the cron job.
// If no cron job is scheduled or no entries are present, it returns the zero time.
func (w *watcher) GetNext() time.Time {
	return w.scheduler.GetNext()
}
