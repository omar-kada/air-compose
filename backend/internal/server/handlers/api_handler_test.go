package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/server/middlewares"
	"omar-kada/air-compose/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
	git.Fetcher
	storage.EventStorage
	storage.DeploymentStorage
}

func (m *Mock) SyncDeployment() (models.Deployment, error) {
	args := m.Called()
	return args.Get(0).(models.Deployment), args.Error(1)
}

func (m *Mock) GetCurrentState() (models.State, error) {
	args := m.Called()
	return args.Get(0).(models.State), args.Error(1)
}

func (m *Mock) GetManagedStacks() (models.StacksState, error) {
	args := m.Called()
	return args.Get(0).(models.StacksState), args.Error(1)
}

func (m *Mock) GetCurrentStacks(services []string) (models.StacksState, error) {
	args := m.Called(services)
	return args.Get(0).(models.StacksState), args.Error(1)
}

func (m *Mock) GetDeployments(c storage.Cursor[uint64]) ([]models.Deployment, error) {
	args := m.Called(c)
	return args.Get(0).([]models.Deployment), args.Error(1)
}

func (m *Mock) GetDeployment(id uint64) (models.Deployment, error) {
	args := m.Called(id)
	return args.Get(0).(models.Deployment), args.Error(1)
}

func (m *Mock) GetEvents(id uint64) ([]models.Event, error) {
	args := m.Called(id)
	return args.Get(0).([]models.Event), args.Error(1)
}

func (m *Mock) GetNotifications(c storage.Cursor[uint64]) ([]models.Event, error) {
	args := m.Called(c)
	return args.Get(0).([]models.Event), args.Error(1)
}

func (m *Mock) GetUser(username string) (models.User, error) {
	args := m.Called(username)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *Mock) DeleteUser(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *Mock) ChangePassword(username string, oldPass, newPass string) (bool, error) {
	args := m.Called(username, oldPass, newPass)
	return args.Bool(0), args.Error(1)
}

func (m *Mock) IsRegistered() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *Mock) TestGitConnection(repo, branch, username, token string) (bool, error) {
	args := m.Called(repo, branch, username, token)
	return args.Bool(0), args.Error(1)
}

func (m *Mock) DiffWithRemote() (models.Patch, error) {
	args := m.Called()
	return args.Get(0).(models.Patch), args.Error(1)
}

type MockStore struct {
	mock.Mock
}

func (m *MockStore) Get() models.Config {
	args := m.Called()
	return args.Get(0).(models.Config)
}

func (m *MockStore) Update(config models.Config) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockStore) ToYaml(config models.Config) ([]byte, error) {
	args := m.Called(config)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStore) SetOnChange(fn func(models.Config, models.Config)) {
	m.Called(fn)
}

func (m *MockStore) WatchFile() error {
	args := m.Called()
	return args.Error(0)
}

func TestDeployementAPIList_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	deps := []models.Deployment{
		{ID: 1, Title: "first", Author: "alice", Diff: "d1", Status: models.DeploymentStatusSuccess},
		{ID: 2, Title: "second", Author: "bob", Diff: "d2", Status: models.DeploymentStatusRunning},
	}
	m.On("GetDeployments", storage.NewIDCursor(2, uint64(0))).Return(deps, nil)

	req := api.DeployementAPIListRequestObject{Params: api.DeployementAPIListParams{Limit: 2}}
	resp, err := h.DeployementAPIList(context.Background(), req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.DeployementAPIList200JSONResponse:
		assert.Equal(t, 2, len(r.Items))
		assert.Equal(t, "2", r.PageInfo.EndCursor)
	default:
		t.Fatalf("unexpected response type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestDeployementAPIList_InvalidOffset(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	off := "notuint"
	req := api.DeployementAPIListRequestObject{Params: api.DeployementAPIListParams{Limit: 1, Offset: &off}}
	resp, err := h.DeployementAPIList(context.Background(), req)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid after value")
}

func TestDeployementAPIRead_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	dep := models.Deployment{ID: 10, Title: "Manual Deploy", Author: "ci", Diff: "diff"}
	m.On("GetDeployment", uint64(10)).Return(dep, nil)
	m.On("GetEvents", uint64(10)).Return([]models.Event{}, nil)

	req := api.DeployementAPIReadRequestObject{Id: "10"}
	resp, err := h.DeployementAPIRead(context.Background(), req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.DeployementAPIRead200JSONResponse:
		assert.Equal(t, "Manual Deploy", r.Title)
		assert.Equal(t, "diff", r.Diff)
	default:
		t.Fatalf("unexpected response type: %T", resp)
	}

	store.AssertExpectations(t)
}

func TestDeployementAPIRead_InvalidID(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	req := api.DeployementAPIReadRequestObject{Id: "abc"}
	resp, err := h.DeployementAPIRead(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestDeployementAPISync_SuccessAndError(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	dep := models.Deployment{ID: 99, Title: "Manual Deploy", Author: "ci", Diff: "dd", Status: models.DeploymentStatusRunning}
	m.On("SyncDeployment").Return(dep, nil)

	resp, err := h.DeployementAPISync(context.Background(), api.DeployementAPISyncRequestObject{})
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.DeployementAPISync200JSONResponse:
		assert.Equal(t, "Manual Deploy", r.Title)
		assert.Equal(t, "dd", r.Diff)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	// now return an error (handler should return both response and error)
	errTest := errors.New("sync failed")
	m.ExpectedCalls = nil
	m.On("SyncDeployment").Return(models.Deployment{}, errTest)

	_, err2 := h.DeployementAPISync(context.Background(), api.DeployementAPISyncRequestObject{})
	assert.Equal(t, errTest, err2)

	m.AssertExpectations(t)
}

func TestStatusAPIGet_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	stacks := models.NewStacksState()
	stacks.SetContainerStatus("service1", models.ContainerSummary{
		ID: "c1", Name: "container1", Image: "img1",
		State: models.StateRunning, Health: models.ContainerHealthy,
	})
	m.On("GetManagedStacks").Return(stacks, nil)

	resp, err := h.StatusAPIGet(context.Background(), api.StatusAPIGetRequestObject{})
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.StatusAPIGet200JSONResponse:
		assert.Equal(t, 1, len(r))
		assert.Contains(t, r, "service1")
		assert.Len(t, r["service1"], 1)
		assert.Equal(t, "container1", r["service1"]["container1"].Name)
		assert.Equal(t, api.ContainerHealthHealthy, r["service1"]["container1"].Health)
		assert.Equal(t, api.ContainerStateRunning, r["service1"]["container1"].State)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestStateAPIGet_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	next := time.Now().Add(1 * time.Hour)
	state := models.State{LastStatus: models.DeploymentStatusError, NextDeploy: next}
	m.On("GetCurrentState").Return(state, nil)

	req := api.StateAPIGetRequestObject{}
	resp, err := h.StateAPIGet(context.Background(), req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.StateAPIGet200JSONResponse:
		assert.Equal(t, next, r.NextDeploy)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestDiffAPIGet_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	fileDiffs := []models.FileDiff{{OldFile: "file1.txt", NewFile: "file1.txt", Diff: "d1"}}
	m.On("DiffWithRemote").Return(models.Patch{
		Files: fileDiffs,
	}, nil)

	resp, err := h.DiffAPIGet(context.Background(), api.DiffAPIGetRequestObject{})
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.DiffAPIGet200JSONResponse:
		assert.Equal(t, 1, len(r))
		assert.Equal(t, "file1.txt", r[0].OldFile)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestDeployementAPIList_GetDeploymentsError(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	errList := errors.New("db error")
	m.On("GetDeployments", storage.NewIDCursor(2, uint64(0))).Return([]models.Deployment{}, errList)

	req := api.DeployementAPIListRequestObject{Params: api.DeployementAPIListParams{Limit: 2}}
	resp, err := h.DeployementAPIList(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, errList, err)

	switch r := resp.(type) {
	case api.DeployementAPIList200JSONResponse:
		assert.Equal(t, 0, len(r.Items))
	default:
		t.Fatalf("unexpected response type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestDeployementAPIList_InvalidLimit(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	req := api.DeployementAPIListRequestObject{Params: api.DeployementAPIListParams{Limit: 0}}
	resp, err := h.DeployementAPIList(context.Background(), req)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid first value")
}

func TestDeployementAPIRead_GetDeploymentError(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	errGet := errors.New("not found")
	m.On("GetDeployment", uint64(10)).Return(models.Deployment{}, errGet)

	req := api.DeployementAPIReadRequestObject{Id: "10"}
	resp, err := h.DeployementAPIRead(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, errGet, err)
	assert.Nil(t, resp)

	store.AssertExpectations(t)
}

func TestStatusAPIGet_Error(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	errStacks := errors.New("failed to get stacks")
	m.On("GetManagedStacks").Return(models.StacksState{}, errStacks)

	resp, err := h.StatusAPIGet(context.Background(), api.StatusAPIGetRequestObject{})
	assert.Nil(t, resp)
	assert.Equal(t, errStacks, err)

	m.AssertExpectations(t)
}

func TestStateAPIGet_Error(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	errState := errors.New("state error")
	m.On("GetCurrentState").Return(models.State{}, errState)

	req := api.StateAPIGetRequestObject{}
	resp, err := h.StateAPIGet(context.Background(), req)
	assert.Nil(t, resp)
	assert.Equal(t, errState, err)

	m.AssertExpectations(t)
}

func TestDiffAPIGet_Error(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	errDiff := errors.New("diff error")
	m.On("DiffWithRemote").Return(models.Patch{
		Files: []models.FileDiff{},
	}, errDiff)

	resp, err := h.DiffAPIGet(context.Background(), api.DiffAPIGetRequestObject{})
	assert.Nil(t, resp)
	assert.Equal(t, errDiff, err)

	m.AssertExpectations(t)
}

func TestConfigAPIGet_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}

	h := NewBusinessHandler(store, m, m, m, m, m, m)

	config := models.Config{
		Environment: models.Environment{
			"ENV": "VALUE",
		},
	}
	store.On("Get").Return(config)

	resp, err := h.ConfigAPIGet(context.Background(), api.ConfigAPIGetRequestObject{})
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.ConfigAPIGet200JSONResponse:
		assert.Equal(t, "VALUE", r.GlobalVariables["ENV"])
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	store.AssertExpectations(t)
}

func TestFeaturesAPIGet_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}

	t.Setenv("AIR_COMPOSE_DISPLAY_CONFIG", "true")
	t.Setenv("AIR_COMPOSE_EDIT_CONFIG", "true")
	t.Setenv("AIR_COMPOSE_EDIT_SETTINGS", "true")

	h := NewBusinessHandler(store, m, m, m, m, m, m)

	resp, err := h.FeaturesAPIGet(context.Background(), api.FeaturesAPIGetRequestObject{})
	assert.NoError(t, err)

	want := api.FeaturesAPIGet200JSONResponse{
		DisplayConfig: true,
		EditConfig:    true,
		EditSettings:  true,
	}

	switch r := resp.(type) {
	case api.FeaturesAPIGet200JSONResponse:
		assert.Equal(t, want, r)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}
}

func TestSettingsAPIGet_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	settings := models.Settings{
		Git: models.GitConfig{

			Repo:     "test-repo",
			Branch:   "main",
			Username: "user",
			Token:    "123456789123456789123456789",
		},
		Schedule: models.ScheduleConfig{
			Cron: "0 0 * * *",
		},
		Notifications: models.NotificationConfig{
			NotificationURL: "gotify://123456789123456789",
		},
	}
	store.On("Get").Return(models.Config{Settings: settings}, nil)

	resp, err := h.SettingsAPIGet(context.Background(), api.SettingsAPIGetRequestObject{})
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.SettingsAPIGet200JSONResponse:
		assert.Equal(t, "test-repo", r.Repo)
		assert.Equal(t, "main", *r.Branch)
		assert.Equal(t, "0 0 * * *", *r.Cron)
		assert.Equal(t, "user", *r.Username)
		assert.Equal(t, "1234567891********************", *r.Token)
		assert.Equal(t, "gotify://1********************", *r.NotificationURL)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	store.AssertExpectations(t)
}

func TestSettingsAPISet_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	oldConfig := models.Config{
		Environment: models.Environment{
			"ENV": "VALUE",
		},
		Services: map[string]models.ServiceConfig{
			"service1": {"key1": "value1"},
		},
		Settings: models.Settings{
			Git: models.GitConfig{

				Repo:     "old-repo",
				Branch:   "old-branch",
				Token:    "123456789",
				Username: "old-user",
			},
			Schedule: models.ScheduleConfig{
				Cron: "old-cron",
			},
			Notifications: models.NotificationConfig{
				NotificationURL: "http://example.com/notification?token=123456",
			},
		},
	}
	newSettings := api.Settings{
		Repo:            "new-repo",
		Branch:          new("new-branch"),
		Cron:            new("new-cron"),
		Username:        new("new-user"),
		Token:           new("******************************"),
		NotificationURL: new("http://ex*********************"),
	}

	store.On("Get").Return(oldConfig, nil)
	store.On("Update", mock.MatchedBy(func(newCfg models.Config) bool {
		// Check that only settings are updated
		assert.Equal(t, oldConfig.Environment, newCfg.Environment)
		assert.Equal(t, oldConfig.Services, newCfg.Services)
		assert.Equal(t, models.Settings{
			Git: models.GitConfig{

				Repo:     newSettings.Repo,
				Branch:   *newSettings.Branch,
				Username: *newSettings.Username,
				Token:    "******************************",
			},
			Schedule: models.ScheduleConfig{

				Cron: *newSettings.Cron,
			},
			Notifications: models.NotificationConfig{
				NotificationURL: "http://ex*********************",
			},
		}, newCfg.Settings)
		return true
	})).Return(nil)

	req := api.SettingsAPISetRequestObject{Body: &newSettings}
	resp, err := h.SettingsAPISet(context.Background(), req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.SettingsAPISet200JSONResponse:
		assert.Equal(t, "new-repo", r.Repo)
		assert.Equal(t, "new-branch", *r.Branch)
		assert.Equal(t, "new-cron", *r.Cron)
		assert.Equal(t, "new-user", *r.Username)
		assert.Equal(t, "******************************", *r.Token)
		assert.Equal(t, "http://ex*********************", *r.NotificationURL)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	store.AssertExpectations(t)
}

func TestSettingsAPISet_UpdateToken(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	oldConfig := models.Config{
		Environment: models.Environment{},
		Services:    map[string]models.ServiceConfig{},
		Settings: models.Settings{
			Git: models.GitConfig{

				Repo:     "old-repo",
				Token:    "123456789",
				Username: "old-user",
			},
		},
	}
	newSettings := api.Settings{
		Repo:     "new-repo",
		Username: new("new-user"),
		Token:    new("123456789123456789123456789"),
	}

	store.On("Get").Return(oldConfig, nil)
	store.On("Update", mock.MatchedBy(func(newCfg models.Config) bool {
		// Check that only settings are updated
		assert.Equal(t, oldConfig.Environment, newCfg.Environment)
		assert.Equal(t, oldConfig.Services, newCfg.Services)
		assert.Equal(t, models.Settings{
			Git: models.GitConfig{

				Repo:     newSettings.Repo,
				Username: *newSettings.Username,
				Token:    *newSettings.Token,
			},
		}, newCfg.Settings)
		return true
	})).Return(nil)

	req := api.SettingsAPISetRequestObject{Body: &newSettings}
	resp, err := h.SettingsAPISet(context.Background(), req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.SettingsAPISet200JSONResponse:
		assert.Equal(t, "new-repo", r.Repo)
		assert.Equal(t, "new-user", *r.Username)
		assert.Equal(t, "1234567891********************", *r.Token)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	store.AssertExpectations(t)
}

func TestSettingsAPISet_Error(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	settings := api.Settings{
		Repo:     "test-repo",
		Branch:   new("main"),
		Cron:     new("0 0 * * *"),
		Token:    new(""),
		Username: new("user"),
	}

	errSettings := errors.New("settings error")
	store.On("Get").Return(models.Config{})
	store.On("Update", mock.Anything).Return(errSettings)

	req := api.SettingsAPISetRequestObject{Body: &settings}
	resp, err := h.SettingsAPISet(context.Background(), req)
	assert.Nil(t, resp)
	assert.Equal(t, errSettings, err)

	store.AssertExpectations(t)
}

func TestConfigAPISet_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	oldConfig := models.Config{
		Environment: models.Environment{
			"ENV": "VALUE",
		},
		Services: map[string]models.ServiceConfig{
			"service1": {"key1": "value1"},
		},
		Settings: models.Settings{
			Git: models.GitConfig{

				Repo:     "old-repo",
				Branch:   "old-branch",
				Username: "old-user",
				Token:    "123456789",
			},
			Schedule: models.ScheduleConfig{
				Cron: "old-cron",
			},
		},
	}
	newConfig := api.Config{
		GlobalVariables: map[string]string{
			"NEW_ENV": "NEW_VALUE",
		},
		Services: map[string]map[string]string{
			"service2": {"key2": "value2"},
		},
	}

	store.On("Get").Return(oldConfig, nil)
	store.On("Update", mock.MatchedBy(func(newCfg models.Config) bool {
		// Check that only environment and services are updated
		assert.Equal(t, models.Environment{"NEW_ENV": "NEW_VALUE"}, newCfg.Environment)
		assert.Equal(t, map[string]models.ServiceConfig{
			"service2": {"key2": "value2"},
		}, newCfg.Services)
		assert.Equal(t, oldConfig.Settings, newCfg.Settings)
		return true
	})).Return(nil)

	req := api.ConfigAPISetRequestObject{Body: &newConfig}
	resp, err := h.ConfigAPISet(context.Background(), req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.ConfigAPISet200JSONResponse:
		assert.Equal(t, "NEW_VALUE", r.GlobalVariables["NEW_ENV"])
		assert.Equal(t, "value2", r.Services["service2"]["key2"])
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	store.AssertExpectations(t)
}

func TestConfigAPISet_Error(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	config := api.Config{
		GlobalVariables: map[string]string{
			"ENV": "VALUE",
		},
		Services: map[string]map[string]string{
			"service1": {"key1": "value1"},
		},
	}

	errConfig := errors.New("config error")
	store.On("Get").Return(models.Config{})
	store.On("Update", mock.Anything).Return(errConfig)

	req := api.ConfigAPISetRequestObject{Body: &config}
	resp, err := h.ConfigAPISet(context.Background(), req)
	assert.Nil(t, resp)
	assert.Equal(t, errConfig, err)

	store.AssertExpectations(t)
}

func TestUserAPIGet_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	user := models.User{Username: "testuser", Type: models.UserTypeLocal}
	ctx := middlewares.ContextWithUsername(context.Background(), user.Username)
	m.On("GetUser", "testuser").Return(user, nil)

	resp, err := h.UserAPIGet(ctx, api.UserAPIGetRequestObject{})
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.UserAPIGet200JSONResponse:
		assert.Equal(t, "testuser", r.Username)
		assert.Equal(t, api.UserTypeLOCAL, r.Type)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}
}

func TestUserAPIGet_NoUser(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	resp, err := h.UserAPIGet(context.Background(), api.UserAPIGetRequestObject{})
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestUserAPIDelete_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	user := models.User{Username: "testuser"}
	ctx := middlewares.ContextWithUsername(context.Background(), user.Username)
	m.On("DeleteUser", "testuser").Return(true, nil)

	resp, err := h.UserAPIDelete(ctx, api.UserAPIDeleteRequestObject{})
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.UserAPIDelete200JSONResponse:
		assert.True(t, r.Success)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestUserAPIDelete_NoUser(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	resp, err := h.UserAPIDelete(context.Background(), api.UserAPIDeleteRequestObject{})
	assert.Error(t, err)
	assert.Equal(t, errUserNotFound, err)

	switch resp.(type) {
	case api.UserAPIDeletedefaultJSONResponse:
		// No specific assertions needed for default response
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}
}

func TestUserAPIDelete_Error(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	user := models.User{Username: "testuser"}
	ctx := middlewares.ContextWithUsername(context.Background(), user.Username)
	errDelete := errors.New("delete error")
	m.On("DeleteUser", "testuser").Return(false, errDelete)

	resp, err := h.UserAPIDelete(ctx, api.UserAPIDeleteRequestObject{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete user")

	switch r := resp.(type) {
	case api.UserAPIDelete200JSONResponse:
		assert.False(t, r.Success)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestAuthAPIRegistered(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	// Mock IsRegistered and Get
	m.On("IsRegistered").Return(true, nil)
	store.On("Get").Return(models.Config{
		Settings: models.Settings{
			Oidc: models.OidcConfig{
				IssuerURL: "https://issuer.example.com",
			},
		},
	}, nil)

	resp, err := h.AuthAPIRegistered(context.Background(), api.AuthAPIRegisteredRequestObject{})
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.AuthAPIRegistered200JSONResponse:
		assert.True(t, r.Registered)
		assert.True(t, r.Oidc)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestAuthAPILogout(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	resp, err := h.AuthAPILogout(context.Background(), api.AuthAPILogoutRequestObject{})
	assert.Error(t, err)
	assert.Equal(t, errShouldntReach, err)

	switch resp.(type) {
	case api.AuthAPILogout200JSONResponse:
		// No specific assertions needed for 200 response
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}
}

func TestAuthAPILogin(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	resp, err := h.AuthAPILogin(context.Background(), api.AuthAPILoginRequestObject{})
	assert.Error(t, err)
	assert.Equal(t, errShouldntReach, err)

	switch resp.(type) {
	case api.AuthAPILogin200JSONResponse:
		// No specific assertions needed for 200 response
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}
}

func TestAuthAPIRegister(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	resp, err := h.AuthAPIRegister(context.Background(), api.AuthAPIRegisterRequestObject{})
	assert.Error(t, err)
	assert.Equal(t, errShouldntReach, err)

	switch resp.(type) {
	case api.AuthAPIRegister200JSONResponse:
		// No specific assertions needed for 200 response
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}
}

func TestUserAPIChangePassword_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	user := models.User{Username: "testuser"}
	ctx := middlewares.ContextWithUsername(context.Background(), user.Username)

	m.On("ChangePassword", "testuser", "oldpass", "newpass").Return(true, nil)

	req := api.UserAPIChangePasswordRequestObject{
		Body: &api.UserAPIChangePasswordJSONRequestBody{
			OldPass: "oldpass",
			NewPass: "newpass",
		},
	}

	resp, err := h.UserAPIChangePassword(ctx, req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.UserAPIChangePassword200JSONResponse:
		assert.True(t, r.Success)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestUserAPIChangePassword_NoUser(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	req := api.UserAPIChangePasswordRequestObject{
		Body: &api.UserAPIChangePasswordJSONRequestBody{
			OldPass: "oldpass",
			NewPass: "newpass",
		},
	}

	resp, err := h.UserAPIChangePassword(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, errUserNotFound, err)

	switch resp.(type) {
	case api.UserAPIChangePassworddefaultJSONResponse:
		// No specific assertions needed for default response
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}
}

func TestUserAPIChangePassword_Error(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	user := models.User{Username: "testuser"}
	ctx := middlewares.ContextWithUsername(context.Background(), user.Username)

	errChange := errors.New("change error")
	m.On("ChangePassword", "testuser", "oldpass", "newpass").Return(false, errChange)

	req := api.UserAPIChangePasswordRequestObject{
		Body: &api.UserAPIChangePasswordJSONRequestBody{
			OldPass: "oldpass",
			NewPass: "newpass",
		},
	}

	resp, err := h.UserAPIChangePassword(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, errChange, err)

	switch r := resp.(type) {
	case api.UserAPIChangePassword200JSONResponse:
		assert.False(t, r.Success)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestSettingsAPITestGitConnection_Success(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	m.On("TestGitConnection", "test-repo", "main", "user", "token").Return(true, nil)

	req := api.SettingsAPITestGitConnectionRequestObject{
		Body: &api.SettingsAPITestGitConnectionJSONRequestBody{
			Repo:     "test-repo",
			Branch:   new("main"),
			Username: new("user"),
			Token:    new("token"),
		},
	}

	resp, err := h.SettingsAPITestGitConnection(context.Background(), req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.SettingsAPITestGitConnection200JSONResponse:
		assert.True(t, r.Success)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestSettingsAPITestGitConnection_Failure(t *testing.T) {
	m := &Mock{}
	store := &MockStore{}
	h := NewBusinessHandler(store, m, m, m, m, m, m)

	errTest := errors.New("connection failed")
	m.On("TestGitConnection", "test-repo", "main", "user", "token").Return(false, errTest)

	req := api.SettingsAPITestGitConnectionRequestObject{
		Body: &api.SettingsAPITestGitConnectionJSONRequestBody{
			Repo:     "test-repo",
			Branch:   new("main"),
			Username: new("user"),
			Token:    new("token"),
		},
	}

	resp, err := h.SettingsAPITestGitConnection(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, errTest, err)

	switch r := resp.(type) {
	case api.SettingsAPITestGitConnection200JSONResponse:
		assert.False(t, r.Success)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}
