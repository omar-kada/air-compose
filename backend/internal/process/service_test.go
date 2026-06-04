package process

import (
	"context"
	"errors"
	"testing"
	"time"

	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/deployments"
	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"

	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/testutil"
	"omar-kada/air-compose/testutil/mocks"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Mocker struct {
	mock.Mock
	mocks.Fetcher
	mocks.Inspector
}

func (m *Mocker) WithCtx(_ context.Context) docker.Deployer {
	return m
}

func (m *Mocker) RemoveServices(services []string, servicesDir string) map[string]error {
	args := m.Called(services, servicesDir)
	return args.Get(0).(map[string]error)
}

func (m *Mocker) DeployServices(cfg models.Config, params models.DeploymentParams) map[string]error {
	args := m.Called(cfg, params)
	return args.Get(0).(map[string]error)
}

func (m *Mocker) RemoveAndDeployStacks(oldCfg, cfg models.Config, params models.DeploymentParams) error {
	args := m.Called(oldCfg, cfg, params)
	return args.Error(0)
}

func (m *Mocker) GetNext() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *Mocker) Schedule(fn func()) (*cron.Cron, error) {
	args := m.Called(fn)
	return args.Get(0).(*cron.Cron), args.Error(1)
}

func (m *Mocker) ReSchedule() (*cron.Cron, error) {
	args := m.Called()
	return args.Get(0).(*cron.Cron), args.Error(1)
}

var (
	mockConfigOld = models.Config{
		Settings: models.Settings{
			Git: models.GitConfig{
				Repo:   "https://example.com/repo.git",
				Branch: "main",
			},
			Notifications: models.NotificationConfig{
				NotificationTypes: []models.EventType{},
			},
		},
		Environment: models.Environment{},
		Services: map[string]models.ServiceConfig{
			"svc1": {
				"Port":    "8080",
				"Version": "v1",
			},
			"svc2": {},
		},
	}
	mockConfigNew = models.Config{
		Settings: models.Settings{
			Git: models.GitConfig{
				Repo:   "https://example.com/repo.git",
				Branch: "main",
			},
			Notifications: models.NotificationConfig{
				NotificationTypes: []models.EventType{},
			},
		},
		Environment: models.Environment{},
		Services: map[string]models.ServiceConfig{
			"svc2": {},
			"svc3": {},
		},
	}
)

func initStore(t *testing.T) deployments.DeploymentStorage {
	db := testutil.NewMemoryStorage(t)
	depStore, err := deployments.NewDeploymentStorage(db)
	if err != nil {
		t.Fatalf("error creating deployment storage : %v", err)
	}
	return depStore
}

func newServiceWithCurrentConfig(t *testing.T, mocker *Mocker, params models.DeploymentParams, currentCfg models.Config) *service {
	configStore, err := config.NewConfigStore(t.TempDir() + "/config.yaml")
	if err != nil {
		t.Fatal("error while creating configStore", err)
	}
	err = configStore.Update(currentCfg)
	if err != nil {
		t.Fatal("error updating config", err)
	}
	depStore := initStore(t)
	svc := NewDeploymentService(
		params,
		mocker,
		mocker,
		depStore,
		configStore,
		events.NewVoidDispatcher(),
	).(*service)
	svc.currentCfg = currentCfg
	return svc
}

func newServiceWithMocks(t *testing.T, mocker *Mocker, params models.DeploymentParams) *service {
	return newServiceWithCurrentConfig(t, mocker, params, models.Config{})
}

var (
	ErrRemove   = errors.New("removeServices error")
	ErrDeploy   = errors.New("deployServices error")
	ErrGenerate = errors.New("generate file error")
	ErrFetch    = errors.New("sync config error")
)

func TestSync_Success(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	})

	wantCfg := mockConfigOld
	err := service.configStore.Update(wantCfg)
	assert.NoError(t, err)
	patch := models.Patch{Diff: "test", Author: "author", CommitHash: "commit"}
	mocker.Fetcher.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mockState := models.NewStacksState()
	mockState.SetContainerStatus("service",
		models.ContainerSummary{
			ID:     "id",
			Name:   "container",
			Health: models.ContainerHealthy,
			State:  models.StateRunning,
		})
	mocker.Inspector.On("GetCurrentStacks", mock.Anything).Return(mockState, nil)
	mocker.On("RemoveAndDeployStacks", models.Config{}, wantCfg, service.params).Once().Return(nil)
	// signal when working branch pull completes
	done := make(chan struct{})
	mocker.Fetcher.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.DoDeploy(DeploymentTriggerConfigurationUpdated, patch)
	assert.NoError(t, err)
	assert.Equal(t, "Configuration changed", dep.Title)
	assert.Equal(t, "test", dep.Diff)
	assert.Equal(t, "author", dep.Author)
	assert.Equal(t, "commit", dep.Commit)
	assert.Equal(t, wantCfg.Settings.Git.Repo, dep.Repo)
	assert.Equal(t, wantCfg.Settings.Git.Branch, dep.Branch)
	assert.Equal(t, models.DeploymentStatusRunning, dep.Status)

	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusSuccess, newDep.Status)

	mocker.AssertExpectations(t)
}

func TestSync_Success_RedploymentWithChangedConfig(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	wantCfg := mockConfigNew
	err := service.configStore.Update(wantCfg)
	assert.NoError(t, err)
	patch := models.Patch{Diff: "test"}
	mockState := models.NewStacksState()
	mockState.SetContainerStatus("service",
		models.ContainerSummary{
			ID:     "id",
			Name:   "container",
			Health: models.ContainerHealthy,
			State:  models.StateRunning,
		})
	mocker.Inspector.On("GetCurrentStacks", mock.Anything).Return(mockState, nil)

	mocker.Fetcher.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", mockConfigOld, wantCfg, service.params).Once().Return(nil)
	// signal when working branch pull completes
	done := make(chan struct{})
	mocker.Fetcher.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })
	dep, err := service.DoDeploy(DeploymentTriggerConfigurationUpdated, patch)
	assert.NoError(t, err)
	assert.Equal(t, "Configuration changed", dep.Title)
	assert.Equal(t, models.DeploymentStatusRunning, dep.Status)

	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusSuccess, newDep.Status)

	mocker.AssertExpectations(t)
}

func TestSync_ErrorsOnPullbranch(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)
	wantCfg := mockConfigNew
	err := service.configStore.Update(wantCfg)
	assert.NoError(t, err)
	patch := models.Patch{Diff: "test"}
	mockState := models.NewStacksState()
	mockState.SetContainerStatus("service",
		models.ContainerSummary{
			ID:     "id",
			Name:   "container",
			Health: models.ContainerHealthy,
			State:  models.StateRunning,
		})
	mocker.Inspector.On("GetCurrentStacks", mock.Anything).Return(mockState, nil)

	done := make(chan struct{})
	mocker.Fetcher.On("PullBranch", WorkingBranch, "").Once().Return(ErrFetch).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.DoDeploy(DeploymentTriggerConfigurationUpdated, patch)
	assert.NoError(t, err)
	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusError, newDep.Status)
	mocker.AssertExpectations(t)
}

func TestSync_Errors(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)
	wantCfg := mockConfigNew
	err := service.configStore.Update(wantCfg)
	assert.NoError(t, err)
	patch := models.Patch{Diff: "test"}
	mockState := models.NewStacksState()
	mockState.SetContainerStatus("service",
		models.ContainerSummary{
			ID:     "id",
			Name:   "container",
			Health: models.ContainerHealthy,
			State:  models.StateRunning,
		})
	mocker.Inspector.On("GetCurrentStacks", mock.Anything).Return(mockState, nil)

	mocker.Fetcher.On("PullBranch", WorkingBranch, "").Once().Return(nil)

	done := make(chan struct{})
	mocker.On("RemoveAndDeployStacks", mockConfigOld, wantCfg, service.params).
		Once().Return(ErrDeploy).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.DoDeploy(DeploymentTriggerConfigurationUpdated, patch)
	assert.NoError(t, err)
	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusError, newDep.Status)
	mocker.AssertExpectations(t)
}

func TestSync_ErrorGettingConfig(t *testing.T) {
	mocker := &Mocker{}
	svc := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	patch := models.Patch{}
	mocker.Fetcher.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", mockConfigOld, mockConfigOld, svc.params).Once().Return(nil)
	done := make(chan struct{})
	mocker.Fetcher.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := svc.DoDeploy(DeploymentTriggerConfigurationUpdated, patch)

	assert.NoError(t, err)
	assert.NotEqual(t, models.Deployment{}, dep)
	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := svc.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusSuccess, newDep.Status)

	mocker.AssertExpectations(t)
}

func TestSync_ConfigNotChanged_StacksHealthy(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	err := service.configStore.Update(mockConfigOld)
	assert.NoError(t, err)
	mockState := models.NewStacksState()
	mockState.SetContainerStatus("service",
		models.ContainerSummary{
			ID:     "id",
			Name:   "container",
			Health: models.ContainerHealthy,
			State:  models.StateRunning,
		})
	mocker.Inspector.On("GetCurrentStacks", mock.Anything).Return(mockState, nil)

	patch := models.Patch{Diff: ""}
	mocker.Fetcher.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", mockConfigOld, mockConfigOld, service.params).Once().Return(nil)
	done := make(chan struct{})
	mocker.Fetcher.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.DoDeploy(DeploymentTriggerConfigurationUpdated, patch)

	assert.NoError(t, err)
	assert.NotEqual(t, models.Deployment{}, dep)
	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusSuccess, newDep.Status)

	mocker.AssertExpectations(t)
}

func TestSync_ConfigNotChanged_StacksUnhealthy(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	err := service.configStore.Update(mockConfigOld)
	assert.NoError(t, err)
	mockState := models.NewStacksState()
	mockState.SetContainerStatus("service",
		models.ContainerSummary{
			ID:     "id",
			Name:   "container",
			Health: models.ContainerUnhealthy,
			State:  models.StateRunning,
		})
	mocker.Inspector.On("GetCurrentStacks", mock.Anything).Return(mockState, nil)

	patch := models.Patch{Diff: ""}
	done := make(chan struct{})
	mocker.Fetcher.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", mockConfigOld, mockConfigOld, service.params).Once().Return(nil)
	mocker.Fetcher.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.DoDeploy(DeploymentTriggerConfigurationUpdated, patch)

	assert.NoError(t, err)
	assert.NotEqual(t, models.Deployment{}, dep)

	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusSuccess, newDep.Status)
	mocker.AssertExpectations(t)
}

func TestSync_NoStacksRunning(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	err := service.configStore.Update(mockConfigOld)
	assert.NoError(t, err)
	mocker.Inspector.On("GetCurrentStacks", mock.Anything).Return(models.NewStacksState(), nil)
	patch := models.Patch{Diff: ""}
	done := make(chan struct{})
	mocker.Fetcher.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", mockConfigOld, mockConfigOld, service.params).Once().Return(nil)
	mocker.Fetcher.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.DoDeploy(DeploymentTriggerConfigurationUpdated, patch)

	assert.NoError(t, err)
	assert.NotEqual(t, models.Deployment{}, dep)

	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusSuccess, newDep.Status)
	mocker.AssertExpectations(t)
}

func TestSync_ErrorCheckingStackHealth(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	err := service.configStore.Update(mockConfigOld)
	assert.NoError(t, err)
	mocker.Inspector.On("GetCurrentStacks", mock.Anything).Return(models.StacksState{}, errors.New("failed to get stacks"))
	patch := models.Patch{Diff: ""}
	done := make(chan struct{})
	mocker.Fetcher.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", mockConfigOld, mockConfigOld, service.params).Once().Return(nil)
	mocker.Fetcher.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.DoDeploy(DeploymentTriggerConfigurationUpdated, patch)

	assert.NoError(t, err)
	assert.NotEqual(t, models.Deployment{}, dep)

	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusSuccess, newDep.Status)

	mocker.AssertExpectations(t)
}

func TestSync_RepositoryAlreadyUpToDate(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	wantCfg := mockConfigOld
	err := service.configStore.Update(wantCfg)
	assert.NoError(t, err)
	mockState := models.NewStacksState()
	mockState.SetContainerStatus("service",
		models.ContainerSummary{
			ID:     "id",
			Name:   "container",
			Health: models.ContainerHealthy,
			State:  models.StateRunning,
		})
	mocker.Inspector.On("GetCurrentStacks", mock.Anything).Return(mockState, nil)

	patch := models.Patch{}
	mocker.Fetcher.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", wantCfg, wantCfg, service.params).Once().Return(nil)
	done := make(chan struct{})
	mocker.Fetcher.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.DoDeploy(DeploymentTriggerConfigurationUpdated, patch)

	assert.NoError(t, err)
	assert.NotEqual(t, models.Deployment{}, dep)
	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusSuccess, newDep.Status)
	mocker.AssertExpectations(t)
}
