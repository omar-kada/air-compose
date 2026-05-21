package process

import (
	"context"
	"errors"
	"testing"
	"time"

	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"
	"omar-kada/air-compose/testutil"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Mocker struct {
	mock.Mock
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

func (m *Mocker) GetManagedStacks() (map[string][]models.ContainerSummary, error) {
	args := m.Called()
	return args.Get(0).(map[string][]models.ContainerSummary), args.Error(1)
}

func (m *Mocker) GetStacksState() (models.StacksState, error) {
	args := m.Called()
	return args.Get(0).(models.StacksState), args.Error(1)
}

func (m *Mocker) GetCurrentStacksState(services []string) (models.StacksState, error) {
	args := m.Called(services)
	return args.Get(0).(models.StacksState), args.Error(1)
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

func (m *Mocker) ClearRepo() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Mocker) CheckoutBranch(branch string) error {
	args := m.Called(branch)
	return args.Error(0)
}

func (m *Mocker) PullBranch(branch string, commitSHA string) error {
	args := m.Called(branch, commitSHA)
	return args.Error(0)
}

func (m *Mocker) DiffWithRemote() (models.Patch, error) {
	args := m.Called()
	return args.Get(0).(models.Patch), args.Error(1)
}

func (m *Mocker) TestGitConnection(repo, branch, username, token string) (bool, error) {
	args := m.Called(repo, branch, username, token)
	return args.Bool(0), args.Error(1)
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

func initStore(t *testing.T) storage.DeploymentStorage {
	db := testutil.NewMemoryStorage(t)
	depStore, err := storage.NewDeploymentStorage(db)
	if err != nil {
		t.Fatalf("error creating deployment storage : %v", err)
	}
	return depStore
}

func newServiceWithCurrentConfig(t *testing.T, mocker *Mocker, params models.DeploymentParams, currentCfg models.Config) *service {
	configStore := storage.NewConfigStore(t.TempDir() + "/config.yaml")
	configStore.Update(currentCfg)
	depStore := initStore(t)
	svc := NewService(
		params,
		mocker,
		mocker,
		mocker,
		depStore,
		configStore,
		events.NewVoidDispatcher(),
		mocker,
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
	service.configStore.Update(wantCfg)
	mocker.On("DiffWithRemote").Once().Return(models.Patch{Diff: "test", Author: "author", CommitHash: "commit"}, nil)
	mocker.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusHealthy,
	}, nil)
	mocker.On("RemoveAndDeployStacks", models.Config{}, wantCfg, service.params).Once().Return(nil)
	// signal when working branch pull completes
	done := make(chan struct{})
	mocker.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.SyncDeployment()
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
	service.configStore.Update(wantCfg)
	mocker.On("DiffWithRemote").Once().Return(models.Patch{Diff: "test"}, nil)
	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusHealthy,
	}, nil)

	mocker.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", mockConfigOld, wantCfg, service.params).Once().Return(nil)
	// signal when working branch pull completes
	done := make(chan struct{})
	mocker.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })
	dep, err := service.SyncDeployment()
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
	service.configStore.Update(wantCfg)
	mocker.On("DiffWithRemote").Once().Return(models.Patch{Diff: "test"}, nil)
	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusHealthy,
	}, nil)

	done := make(chan struct{})
	mocker.On("PullBranch", WorkingBranch, "").Once().Return(ErrFetch).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.SyncDeployment()
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
	service.configStore.Update(wantCfg)
	mocker.On("DiffWithRemote").Once().Return(models.Patch{Diff: "test"}, nil)
	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusHealthy,
	}, nil)

	mocker.On("PullBranch", WorkingBranch, "").Once().Return(nil)

	done := make(chan struct{})
	mocker.On("RemoveAndDeployStacks", mockConfigOld, wantCfg, service.params).
		Once().Return(ErrDeploy).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.SyncDeployment()
	assert.NoError(t, err)
	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusError, newDep.Status)
	mocker.AssertExpectations(t)
}

func TestGetCurrentState_NoDeployments(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{})

	next := time.Now().Add(1 * time.Hour)
	mocker.On("GetNext").Return(next)
	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusHealthy,
	}, nil)

	state, err := service.GetCurrentState()
	assert.NoError(t, err)
	assert.Equal(t, next, state.NextDeploy)
	mocker.AssertExpectations(t)
}

func TestGetCurrentState_WithDeployments(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{})

	next := time.Now().Add(30 * time.Minute)
	mocker.On("GetNext").Return(next)
	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusHealthy,
	}, nil)

	// create a successful deployment
	dep1, err := service.store.InitDeployment("first", models.Patch{}, models.GitConfig{})
	assert.NoError(t, err)
	err = service.store.EndDeployment(dep1.ID, models.DeploymentStatusSuccess)
	assert.NoError(t, err)

	// create a failed (last) deployment
	dep2, err := service.store.InitDeployment("second", models.Patch{}, models.GitConfig{})
	assert.NoError(t, err)
	err = service.store.EndDeployment(dep2.ID, models.DeploymentStatusError)
	assert.NoError(t, err)
	state, err := service.GetCurrentState()
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusError, state.LastStatus)
	assert.Equal(t, next, state.NextDeploy)
	mocker.AssertExpectations(t)
}

func TestSync_ErrorGettingConfig(t *testing.T) {
	mocker := &Mocker{}
	configStore := storage.NewConfigStore(t.TempDir() + "/config.yaml")
	depStore := initStore(t)

	// Don't initialize config, so Get() will fail
	svc := NewService(
		models.DeploymentParams{
			ServicesDir: "/services",
			WorkingDir:  ".",
		},
		mocker,
		mocker,
		mocker,
		depStore,
		configStore,
		events.NewVoidDispatcher(),
		mocker,
	).(*service)

	dep, err := svc.SyncDeployment()

	assert.ErrorContains(t, err, "error getting repo")
	assert.Equal(t, models.Deployment{}, dep)
	mocker.AssertExpectations(t)
}

func TestSync_ConfigNotChanged_StacksHealthy(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	service.configStore.Update(mockConfigOld)

	/*healthyContainer := models.ContainerSummary{
		ID:     "container1",
		Name:   "container1",
		Image:  "image1",
		State:  container.StateRunning,
		Health: container.Healthy,
	}*/
	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusHealthy,
	}, nil)

	mocker.On("DiffWithRemote").Return(models.Patch{Diff: ""}, nil)

	dep, err := service.SyncDeployment()

	assert.NoError(t, err)
	assert.Equal(t, models.Deployment{}, dep)
	mocker.AssertExpectations(t)
}

func TestSync_ConfigNotChanged_StacksUnhealthy(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	service.configStore.Update(mockConfigOld)

	// unhealthyContainer := models.ContainerSummary{
	// 	ID:     "container1",
	// 	Name:   "container1",
	// 	Image:  "image1",
	// 	State:  container.StateRunning,
	// 	Health: container.Unhealthy,
	// }
	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusUnhealthy,
	}, nil)

	mocker.On("DiffWithRemote").Return(models.Patch{Diff: ""}, nil)
	done := make(chan struct{})
	mocker.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", mockConfigOld, mockConfigOld, service.params).Once().Return(nil)
	mocker.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.SyncDeployment()

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

	service.configStore.Update(mockConfigOld)

	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusUnknown,
	}, nil)
	mocker.On("DiffWithRemote").Return(models.Patch{Diff: ""}, nil)
	done := make(chan struct{})
	mocker.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", mockConfigOld, mockConfigOld, service.params).Once().Return(nil)
	mocker.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.SyncDeployment()

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

	service.configStore.Update(mockConfigOld)

	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{}, errors.New("failed to get stacks"))
	mocker.On("DiffWithRemote").Return(models.Patch{Diff: ""}, nil)
	done := make(chan struct{})
	mocker.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("RemoveAndDeployStacks", mockConfigOld, mockConfigOld, service.params).Once().Return(nil)
	mocker.On("PullBranch", "main", mock.Anything).Once().
		Return(nil).
		Run(func(_ mock.Arguments) { close(done) })

	dep, err := service.SyncDeployment()

	assert.NoError(t, err)
	assert.NotEqual(t, models.Deployment{}, dep)

	testutil.WaitForChannel(t, done, 1*time.Second, "timeout waiting for background deployment goroutine")
	time.Sleep(10 * time.Millisecond)

	newDep, err := service.store.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusSuccess, newDep.Status)

	mocker.AssertExpectations(t)
}

func TestGetCurrentState_ErrorGettingStacks(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{})

	next := time.Now().Add(1 * time.Hour)
	mocker.On("GetNext").Return(next)
	mocker.On("GetCurrentStacksState", mock.Anything).Return(
		models.NewStacksState(), errors.New("failed to get stacks"))

	state, err := service.GetCurrentState()

	assert.NoError(t, err)
	assert.Equal(t, models.State{
		NextDeploy: next,
		Health:     models.StackStatusUnknown,
	}, state)

	mocker.AssertExpectations(t)
}

func TestGetCurrentState_MultipleDeploymentsVariousStatuses(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{})

	next := time.Now().Add(30 * time.Minute)
	mocker.On("GetNext").Return(next)

	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusUnknown,
	}, nil)

	// Create multiple deployments
	dep1, _ := service.store.InitDeployment("first", models.Patch{}, models.GitConfig{})
	service.store.EndDeployment(dep1.ID, models.DeploymentStatusSuccess)

	dep2, _ := service.store.InitDeployment("second", models.Patch{}, models.GitConfig{})
	service.store.EndDeployment(dep2.ID, models.DeploymentStatusSuccess)

	dep3, _ := service.store.InitDeployment("third", models.Patch{}, models.GitConfig{})
	service.store.EndDeployment(dep3.ID, models.DeploymentStatusError)

	dep4, _ := service.store.InitDeployment("fourth", models.Patch{}, models.GitConfig{})
	service.store.EndDeployment(dep4.ID, models.DeploymentStatusError)

	state, err := service.GetCurrentState()

	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusError, state.LastStatus)
	assert.Equal(t, next, state.NextDeploy)
	mocker.AssertExpectations(t)
}

func TestSync_RepositoryAlreadyUpToDate(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	wantCfg := mockConfigOld
	service.configStore.Update(wantCfg)
	// healthyContainer := models.ContainerSummary{
	// 	ID:     "container1",
	// 	Name:   "container1",
	// 	Image:  "image1",
	// 	State:  container.StateRunning,
	// 	Health: container.Healthy,
	// }
	mocker.On("GetCurrentStacksState", mock.Anything).Return(models.StacksState{
		GlobalStatus: models.StackStatusHealthy,
	}, nil)

	mocker.On("DiffWithRemote").Once().Return(models.Patch{}, git.NoErrAlreadyUpToDate)

	dep, err := service.SyncDeployment()

	assert.NoError(t, err)
	assert.Equal(t, models.Deployment{}, dep)
	mocker.AssertExpectations(t)
}
