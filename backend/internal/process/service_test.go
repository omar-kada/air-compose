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

	"github.com/moby/moby/api/types/container"
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

func (m *Mocker) GetManagedStacks(servicesDir string) (map[string][]models.ContainerSummary, error) {
	args := m.Called(servicesDir)
	return args.Get(0).(map[string][]models.ContainerSummary), args.Error(1)
}

func (m *Mocker) GetServiceContainers(serviceName string, servicesDir string) ([]string, error) {
	args := m.Called(serviceName, servicesDir)
	return args.Get(0).([]string), args.Error(1)
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

func (m *Mocker) DiffWithRemote() (git.Patch, error) {
	args := m.Called()
	return args.Get(0).(git.Patch), args.Error(1)
}

var (
	mockConfigOld = models.Config{
		Settings: models.Settings{
			Repo:              "https://example.com/repo.git",
			Branch:            "main",
			NotificationTypes: []models.EventType{},
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
			Repo:              "https://example.com/repo.git",
			Branch:            "main",
			NotificationTypes: []models.EventType{},
		},
		Environment: models.Environment{},
		Services: map[string]models.ServiceConfig{
			"svc2": {},
			"svc3": {},
		},
	}
)

func initStores(t *testing.T) (storage.DeploymentStorage, storage.EventStorage) {
	db := testutil.NewMemoryStorage()
	depStore, err := storage.NewDeploymentStorage(db)
	if err != nil {
		t.Fatalf("error creating deployment storage : %v", err)
	}
	eventStore, err := storage.NewEventStorage(db)
	if err != nil {
		t.Fatalf("error creating deployment storage : %v", err)
	}
	return depStore, eventStore
}

func newServiceWithCurrentConfig(t *testing.T, mocker *Mocker, params models.DeploymentParams, currentCfg models.Config) *service {
	configStore := storage.NewConfigStore(t.TempDir() + "/config.yaml")
	configStore.Update(currentCfg)
	depStore, eventStore := initStores(t)
	svc := NewService(
		params,
		mocker,
		mocker,
		mocker,
		depStore, eventStore,
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
	mocker.On("DiffWithRemote").Once().Return(git.Patch{Diff: "test"}, nil)
	mocker.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("GetManagedStacks", mock.Anything).Return(map[string][]models.ContainerSummary{}, nil)
	mocker.On("GetServiceContainers", mock.Anything, mock.Anything).Return([]string{"container"}, nil)
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
	mocker.On("DiffWithRemote").Once().Return(git.Patch{Diff: "test"}, nil)
	mocker.On("GetManagedStacks", mock.Anything).Return(map[string][]models.ContainerSummary{}, nil)
	mocker.On("GetServiceContainers", mock.Anything, mock.Anything).Return([]string{"container"}, nil)

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
	mocker.On("DiffWithRemote").Once().Return(git.Patch{Diff: "test"}, nil)
	mocker.On("GetManagedStacks", mock.Anything).Return(map[string][]models.ContainerSummary{}, nil)
	mocker.On("GetServiceContainers", mock.Anything, mock.Anything).Return([]string{"container"}, nil)

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
}

func TestSync_Errors(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)
	wantCfg := mockConfigNew
	service.configStore.Update(wantCfg)
	mocker.On("DiffWithRemote").Once().Return(git.Patch{Diff: "test"}, nil)
	mocker.On("GetManagedStacks", mock.Anything).Return(map[string][]models.ContainerSummary{}, nil)
	mocker.On("GetServiceContainers", mock.Anything, mock.Anything).Return([]string{"container"}, nil)

	mocker.On("PullBranch", WorkingBranch, "").Once().Return(nil)
	mocker.On("CheckoutBranch", "main").Once().Return(nil)

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
}

func TestGetCurrentState_NoDeployments(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{})

	next := time.Now().Add(1 * time.Hour)
	mocker.On("GetNext").Return(next)
	mocker.On("GetManagedStacks", mock.Anything).Return(map[string][]models.ContainerSummary{}, nil)

	state, err := service.GetCurrentState()
	assert.NoError(t, err)
	assert.Equal(t, next, state.NextDeploy)
}

func TestGetCurrentState_WithDeployments(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{})

	next := time.Now().Add(30 * time.Minute)
	mocker.On("GetNext").Return(next)
	mocker.On("GetManagedStacks", mock.Anything).Return(map[string][]models.ContainerSummary{}, nil)

	// create a successful deployment
	dep1, err := service.store.InitDeployment("first", "alice", "diff1", nil)
	assert.NoError(t, err)
	err = service.store.EndDeployment(dep1.ID, models.DeploymentStatusSuccess)
	assert.NoError(t, err)

	// create a failed (last) deployment
	dep2, err := service.store.InitDeployment("second", "bob", "diff2", nil)
	assert.NoError(t, err)
	err = service.store.EndDeployment(dep2.ID, models.DeploymentStatusError)
	assert.NoError(t, err)
	state, err := service.GetCurrentState()
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusError, state.LastStatus)
	assert.Equal(t, next, state.NextDeploy)
}

func TestGetManagedStacks(t *testing.T) {
	testCases := []struct {
		name           string
		servicesDir    string
		expectedResult map[string][]models.ContainerSummary
		expectedError  error
	}{
		{
			name:        "Success",
			servicesDir: "/services",
			expectedResult: map[string][]models.ContainerSummary{
				"service1": {
					{
						ID:     "container1",
						Name:   "container1",
						Image:  "image1",
						State:  container.StateRunning,
						Health: container.Healthy,
					},
				},
			},
			expectedError: nil,
		},
		{
			name:           "Error",
			servicesDir:    "/services",
			expectedResult: nil,
			expectedError:  errors.New("failed to get managed containers"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mocker := &Mocker{}
			service := newServiceWithMocks(t, mocker, models.DeploymentParams{
				ServicesDir: tc.servicesDir,
			})

			mocker.On("GetManagedStacks", tc.servicesDir).Return(tc.expectedResult, tc.expectedError)

			result, err := service.GetManagedStacks()

			if tc.expectedError != nil {
				assert.ErrorContains(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestGetDiff_Success(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{}, mockConfigOld)

	expectedDiff := []models.FileDiff{
		{OldFile: "file1.txt", NewFile: "file1.txt", Diff: "diff1"},
		{OldFile: "file2.txt", NewFile: "file2.txt", Diff: "diff2"},
	}

	mocker.On("DiffWithRemote").Return(git.Patch{Files: expectedDiff}, nil)

	diff, err := service.GetDiff()

	assert.NoError(t, err)
	assert.Equal(t, expectedDiff, diff)
	mocker.AssertExpectations(t)
}

func TestGetDiff_ErrorDiffWithRemote(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{}, mockConfigOld)

	mocker.On("DiffWithRemote").Return(git.Patch{}, ErrFetch)

	diff, err := service.GetDiff()

	assert.ErrorContains(t, err, "sync config error")
	assert.Equal(t, 0, len(diff))
	mocker.AssertExpectations(t)
}

// Edge case tests

func TestSync_ErrorGettingConfig(t *testing.T) {
	mocker := &Mocker{}
	configStore := storage.NewConfigStore(t.TempDir() + "/config.yaml")
	depStore, eventStore := initStores(t)

	// Don't initialize config, so Get() will fail
	svc := NewService(
		models.DeploymentParams{
			ServicesDir: "/services",
			WorkingDir:  ".",
		},
		mocker,
		mocker,
		mocker,
		depStore, eventStore,
		configStore,
		events.NewVoidDispatcher(),
		mocker,
	).(*service)

	dep, err := svc.SyncDeployment()

	assert.ErrorContains(t, err, "error getting repo")
	assert.Equal(t, models.Deployment{}, dep)
}

func TestSync_ConfigNotChanged_StacksHealthy(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	service.configStore.Update(mockConfigOld)

	healthyContainer := models.ContainerSummary{
		ID:     "container1",
		Name:   "container1",
		Image:  "image1",
		State:  container.StateRunning,
		Health: container.Healthy,
	}
	mocker.On("GetManagedStacks", "/services").Return(map[string][]models.ContainerSummary{
		"svc1": {healthyContainer},
		"svc2": {healthyContainer},
	}, nil)
	mocker.On("GetServiceContainers", mock.Anything, mock.Anything).Return([]string{"container1"}, nil)

	mocker.On("DiffWithRemote").Return(git.Patch{Diff: ""}, nil)

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

	unhealthyContainer := models.ContainerSummary{
		ID:     "container1",
		Name:   "container1",
		Image:  "image1",
		State:  container.StateRunning,
		Health: container.Unhealthy,
	}
	mocker.On("GetManagedStacks", "/services").Return(map[string][]models.ContainerSummary{
		"svc1": {unhealthyContainer},
	}, nil)
	mocker.On("GetServiceContainers", mock.Anything, mock.Anything).Return([]string{"container1"}, nil)

	mocker.On("DiffWithRemote").Return(git.Patch{Diff: ""}, nil)
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

	mocker.On("GetManagedStacks", "/services").Return(map[string][]models.ContainerSummary{}, nil)
	mocker.On("GetServiceContainers", mock.Anything, mock.Anything).Return([]string{"container1"}, nil)
	mocker.On("DiffWithRemote").Return(git.Patch{Diff: ""}, nil)
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
}

func TestSync_ErrorCheckingStackHealth(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	service.configStore.Update(mockConfigOld)

	mocker.On("GetManagedStacks", "/services").Return(map[string][]models.ContainerSummary{}, errors.New("failed to get stacks"))
	mocker.On("DiffWithRemote").Return(git.Patch{Diff: ""}, nil)
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
}

func TestGetCurrentState_ErrorGettingStacks(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{})

	next := time.Now().Add(1 * time.Hour)
	mocker.On("GetNext").Return(next)
	mocker.On("GetManagedStacks", mock.Anything).Return(map[string][]models.ContainerSummary{}, errors.New("failed to get stacks"))

	state, err := service.GetCurrentState()

	assert.NoError(t, err)
	assert.Equal(t, models.State{
		NextDeploy: next,
		Health:     models.StackStatusUnknown,
	}, state)
}

func TestGetCurrentState_MultipleDeploymentsVariousStatuses(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{})

	next := time.Now().Add(30 * time.Minute)
	mocker.On("GetNext").Return(next)
	mocker.On("GetManagedStacks", mock.Anything).Return(map[string][]models.ContainerSummary{}, nil)

	// Create multiple deployments
	dep1, _ := service.store.InitDeployment("first", "alice", "diff1", nil)
	service.store.EndDeployment(dep1.ID, models.DeploymentStatusSuccess)

	dep2, _ := service.store.InitDeployment("second", "bob", "diff2", nil)
	service.store.EndDeployment(dep2.ID, models.DeploymentStatusSuccess)

	dep3, _ := service.store.InitDeployment("third", "charlie", "diff3", nil)
	service.store.EndDeployment(dep3.ID, models.DeploymentStatusError)

	dep4, _ := service.store.InitDeployment("fourth", "david", "diff4", nil)
	service.store.EndDeployment(dep4.ID, models.DeploymentStatusError)

	state, err := service.GetCurrentState()

	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusError, state.LastStatus)
	assert.Equal(t, next, state.NextDeploy)
}

func TestGetDiff_NoCurrentConfigUsingConfigStore(t *testing.T) {
	mocker := &Mocker{}
	configStore := storage.NewConfigStore(t.TempDir() + "/config.yaml")
	configStore.Update(mockConfigOld)
	depStore, eventStore := initStores(t)

	svc := NewService(
		models.DeploymentParams{},
		mocker,
		mocker,
		mocker,
		depStore, eventStore,
		configStore,
		events.NewVoidDispatcher(),
		mocker,
	).(*service)

	expectedDiff := []models.FileDiff{
		{OldFile: "file1.txt", NewFile: "file1.txt", Diff: "diff1"},
	}

	mocker.On("DiffWithRemote").Return(git.Patch{Files: expectedDiff}, nil)

	diff, err := svc.GetDiff()

	assert.NoError(t, err)
	assert.Equal(t, expectedDiff, diff)
	mocker.AssertExpectations(t)
}

func TestGetDeployments_Success(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{})

	dep1, _ := service.store.InitDeployment("first", "alice", "diff1", nil)
	service.store.EndDeployment(dep1.ID, models.DeploymentStatusSuccess)

	dep2, _ := service.store.InitDeployment("second", "bob", "diff2", nil)
	service.store.EndDeployment(dep2.ID, models.DeploymentStatusError)

	deployments, err := service.GetDeployments(10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(deployments))
	assert.Equal(t, "second", deployments[0].Title)
	assert.Equal(t, "first", deployments[1].Title)
}

func TestGetDeployments_WithPagination(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithMocks(t, mocker, models.DeploymentParams{})

	for i := 1; i <= 6; i++ {
		dep, _ := service.store.InitDeployment("deployment"+string(rune(i)), "author", "diff", nil)
		service.store.EndDeployment(dep.ID, models.DeploymentStatusSuccess)
	}

	// Get first page
	page1, err := service.GetDeployments(2, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(page1))

	// Get second page
	page2, err := service.GetDeployments(2, 5)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(page2))

	// Get third page
	page3, err := service.GetDeployments(2, 3)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(page3))
}

func TestSync_RepositoryAlreadyUpToDate(t *testing.T) {
	mocker := &Mocker{}
	service := newServiceWithCurrentConfig(t, mocker, models.DeploymentParams{
		ServicesDir: "/services",
		WorkingDir:  ".",
	}, mockConfigOld)

	wantCfg := mockConfigOld
	service.configStore.Update(wantCfg)
	healthyContainer := models.ContainerSummary{
		ID:     "container1",
		Name:   "container1",
		Image:  "image1",
		State:  container.StateRunning,
		Health: container.Healthy,
	}
	mocker.On("GetManagedStacks", "/services").Return(map[string][]models.ContainerSummary{
		"svc1": {healthyContainer},
		"svc2": {healthyContainer},
	}, nil)
	mocker.On("GetServiceContainers", mock.Anything, mock.Anything).Return([]string{"container1"}, nil)

	mocker.On("DiffWithRemote").Once().Return(git.Patch{}, git.NoErrAlreadyUpToDate)

	dep, err := service.SyncDeployment()

	assert.NoError(t, err)
	assert.Equal(t, models.Deployment{}, dep)
	mocker.AssertExpectations(t)
}
