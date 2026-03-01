package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/server/middlewares"
	"omar-kada/air-compose/models"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProcess struct {
	mock.Mock
}

func (m *MockProcess) SyncDeployment() (models.Deployment, error) {
	args := m.Called()
	return args.Get(0).(models.Deployment), args.Error(1)
}

func (m *MockProcess) GetCurrentStats(days int) (models.Stats, error) {
	args := m.Called(days)
	return args.Get(0).(models.Stats), args.Error(1)
}

func (m *MockProcess) GetDiff() ([]models.FileDiff, error) {
	args := m.Called()
	return args.Get(0).([]models.FileDiff), args.Error(1)
}

func (m *MockProcess) GetManagedStacks() (map[string][]models.ContainerSummary, error) {
	args := m.Called()
	return args.Get(0).(map[string][]models.ContainerSummary), args.Error(1)
}

func (m *MockProcess) GetDeployments(limit int, offset uint64) ([]models.Deployment, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.Deployment), args.Error(1)
}

func (m *MockProcess) GetDeployment(id uint64) (models.Deployment, error) {
	args := m.Called(id)
	return args.Get(0).(models.Deployment), args.Error(1)
}

func (m *MockProcess) GetNotifications(limit int, offset uint64) ([]models.Event, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.Event), args.Error(1)
}

func (m *MockProcess) GetUser(username string) (models.User, error) {
	args := m.Called(username)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockProcess) DeleteUser(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MockProcess) ChangePassword(username string, oldPass, newPass string) (bool, error) {
	args := m.Called(username, oldPass, newPass)
	return args.Bool(0), args.Error(1)
}

type MockStore struct {
	mock.Mock
}

func (m *MockStore) Get() (models.Config, error) {
	args := m.Called()
	return args.Get(0).(models.Config), args.Error(1)
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

func TestDeployementAPIList_Success(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	deps := []models.Deployment{
		{ID: 1, Title: "first", Author: "alice", Diff: "d1", Status: models.DeploymentStatusSuccess},
		{ID: 2, Title: "second", Author: "bob", Diff: "d2", Status: models.DeploymentStatusRunning},
	}
	m.On("GetDeployments", 2, uint64(0)).Return(deps, nil)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	off := "notuint"
	req := api.DeployementAPIListRequestObject{Params: api.DeployementAPIListParams{Limit: 1, Offset: &off}}
	resp, err := h.DeployementAPIList(context.Background(), req)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid after value")
}

func TestDeployementAPIRead_Success(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	dep := models.Deployment{ID: 10, Title: "Manual Deploy", Author: "ci", Diff: "diff"}
	m.On("GetDeployment", uint64(10)).Return(dep, nil)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	req := api.DeployementAPIReadRequestObject{Id: "abc"}
	resp, err := h.DeployementAPIRead(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestDeployementAPISync_SuccessAndError(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	stacks := map[string][]models.ContainerSummary{
		"stack1": {{ID: "c1", Name: "c1", Image: "img1", State: container.StateRunning, Health: container.Healthy}},
	}
	m.On("GetManagedStacks").Return(stacks, nil)

	resp, err := h.StatusAPIGet(context.Background(), api.StatusAPIGetRequestObject{})
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.StatusAPIGet200JSONResponse:
		assert.Equal(t, 1, len(r))
		assert.Equal(t, "stack1", r[0].StackId)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestStatsAPIGet_Success(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	next := time.Now().Add(1 * time.Hour)
	stats := models.Stats{Author: "bob", Error: 1, Success: 2, LastStatus: models.DeploymentStatusError, NextDeploy: next}
	m.On("GetCurrentStats", 7).Return(stats, nil)

	req := api.StatsAPIGetRequestObject{Days: 7}
	resp, err := h.StatsAPIGet(context.Background(), req)
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.StatsAPIGet200JSONResponse:
		assert.Equal(t, int32(2), r.Success)
		assert.Equal(t, int32(1), r.Error)
		assert.Equal(t, "bob", r.Author)
		assert.Equal(t, next, r.NextDeploy)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}

	m.AssertExpectations(t)
}

func TestDiffAPIGet_Success(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	fileDiffs := []models.FileDiff{{OldFile: "file1.txt", NewFile: "file1.txt", Diff: "d1"}}
	m.On("GetDiff").Return(fileDiffs, nil)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	errList := errors.New("db error")
	m.On("GetDeployments", 2, uint64(0)).Return([]models.Deployment{}, errList)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	req := api.DeployementAPIListRequestObject{Params: api.DeployementAPIListParams{Limit: 0}}
	resp, err := h.DeployementAPIList(context.Background(), req)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "invalid first value")
}

func TestDeployementAPIRead_GetDeploymentError(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	errStacks := errors.New("failed to get stacks")
	m.On("GetManagedStacks").Return(map[string][]models.ContainerSummary{}, errStacks)

	resp, err := h.StatusAPIGet(context.Background(), api.StatusAPIGetRequestObject{})
	assert.Nil(t, resp)
	assert.Equal(t, errStacks, err)

	m.AssertExpectations(t)
}

func TestStatsAPIGet_Error(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	errStats := errors.New("stats error")
	m.On("GetCurrentStats", 7).Return(models.Stats{}, errStats)

	req := api.StatsAPIGetRequestObject{Days: 7}
	resp, err := h.StatsAPIGet(context.Background(), req)
	assert.Nil(t, resp)
	assert.Equal(t, errStats, err)

	m.AssertExpectations(t)
}

func TestDiffAPIGet_Error(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	errDiff := errors.New("diff error")
	m.On("GetDiff").Return([]models.FileDiff{}, errDiff)

	resp, err := h.DiffAPIGet(context.Background(), api.DiffAPIGetRequestObject{})
	assert.Nil(t, resp)
	assert.Equal(t, errDiff, err)

	m.AssertExpectations(t)
}

func TestConfigAPIGet_Success(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}

	h := NewHandler(store, m, m)

	config := models.Config{
		Environment: models.Environment{
			"ENV": "VALUE",
		},
	}
	store.On("Get").Return(config, nil)

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

func TestConfigAPIGet_Error(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	errConfig := errors.New("config error")
	store.On("Get").Return(models.Config{}, errConfig)

	resp, err := h.ConfigAPIGet(context.Background(), api.ConfigAPIGetRequestObject{})
	assert.Nil(t, resp)
	assert.Equal(t, errConfig, err)

	store.AssertExpectations(t)
}

func TestFeaturesAPIGet_Success(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}

	t.Setenv("AIR_COMPOSE_DISPLAY_CONFIG", "true")
	t.Setenv("AIR_COMPOSE_EDIT_CONFIG", "true")
	t.Setenv("AIR_COMPOSE_EDIT_SETTINGS", "true")

	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	settings := models.Settings{
		Repo:            "test-repo",
		Branch:          "main",
		Cron:            "0 0 * * *",
		Username:        "user",
		Token:           "123456789123456789123456789",
		NotificationURL: "gotify://123456789123456789",
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

func TestSettingsAPIGet_Error(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	errSettings := errors.New("settings error")
	store.On("Get").Return(models.Config{}, errSettings)

	resp, err := h.SettingsAPIGet(context.Background(), api.SettingsAPIGetRequestObject{})
	assert.Nil(t, resp)
	assert.Equal(t, errSettings, err)

	store.AssertExpectations(t)
}

func TestSettingsAPISet_Success(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	oldConfig := models.Config{
		Environment: models.Environment{
			"ENV": "VALUE",
		},
		Services: map[string]models.ServiceConfig{
			"service1": {"key1": "value1"},
		},
		Settings: models.Settings{
			Repo:            "old-repo",
			Branch:          "old-branch",
			Cron:            "old-cron",
			Token:           "123456789",
			NotificationURL: "http://example.com/notification?token=123456",
			Username:        "old-user",
		},
	}
	newSettings := api.Settings{
		Repo:            "new-repo",
		Branch:          ptr("new-branch"),
		Cron:            ptr("new-cron"),
		Username:        ptr("new-user"),
		Token:           ptr("******************************"),
		NotificationURL: ptr("http://ex*********************"),
	}

	store.On("Get").Return(oldConfig, nil)
	store.On("Update", mock.MatchedBy(func(newCfg models.Config) bool {
		// Check that only settings are updated
		assert.Equal(t, oldConfig.Environment, newCfg.Environment)
		assert.Equal(t, oldConfig.Services, newCfg.Services)
		assert.Equal(t, models.Settings{
			Repo:            newSettings.Repo,
			Branch:          *newSettings.Branch,
			Cron:            *newSettings.Cron,
			Username:        *newSettings.Username,
			Token:           "******************************",
			NotificationURL: "http://ex*********************",
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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	oldConfig := models.Config{
		Environment: models.Environment{},
		Services:    map[string]models.ServiceConfig{},
		Settings: models.Settings{
			Repo:     "old-repo",
			Token:    "123456789",
			Username: "old-user",
		},
	}
	newSettings := api.Settings{
		Repo:     "new-repo",
		Username: ptr("new-user"),
		Token:    ptr("123456789123456789123456789"),
	}

	store.On("Get").Return(oldConfig, nil)
	store.On("Update", mock.MatchedBy(func(newCfg models.Config) bool {
		// Check that only settings are updated
		assert.Equal(t, oldConfig.Environment, newCfg.Environment)
		assert.Equal(t, oldConfig.Services, newCfg.Services)
		assert.Equal(t, models.Settings{
			Repo:     newSettings.Repo,
			Username: *newSettings.Username,
			Token:    *newSettings.Token,
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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	settings := api.Settings{
		Repo:     "test-repo",
		Branch:   ptr("main"),
		Cron:     ptr("0 0 * * *"),
		Token:    ptr(""),
		Username: ptr("user"),
	}

	errSettings := errors.New("settings error")
	store.On("Get").Return(models.Config{}, errSettings)

	req := api.SettingsAPISetRequestObject{Body: &settings}
	resp, err := h.SettingsAPISet(context.Background(), req)
	assert.Nil(t, resp)
	assert.Equal(t, errSettings, err)

	store.AssertExpectations(t)
}

func ptr[T any](v T) *T {
	return &v
}

func TestConfigAPISet_Success(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	oldConfig := models.Config{
		Environment: models.Environment{
			"ENV": "VALUE",
		},
		Services: map[string]models.ServiceConfig{
			"service1": {"key1": "value1"},
		},
		Settings: models.Settings{
			Repo:     "old-repo",
			Branch:   "old-branch",
			Cron:     "old-cron",
			Token:    "123456789",
			Username: "old-user",
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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	config := api.Config{
		GlobalVariables: map[string]string{
			"ENV": "VALUE",
		},
		Services: map[string]map[string]string{
			"service1": {"key1": "value1"},
		},
	}

	errConfig := errors.New("config error")
	store.On("Get").Return(models.Config{}, errConfig)

	req := api.ConfigAPISetRequestObject{Body: &config}
	resp, err := h.ConfigAPISet(context.Background(), req)
	assert.Nil(t, resp)
	assert.Equal(t, errConfig, err)

	store.AssertExpectations(t)
}

func TestUserAPIGet_Success(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	user := models.User{Username: "testuser"}
	ctx := middlewares.ContextWithUsername(context.Background(), user.Username)

	resp, err := h.UserAPIGet(ctx, api.UserAPIGetRequestObject{})
	assert.NoError(t, err)

	switch r := resp.(type) {
	case api.UserAPIGet200JSONResponse:
		assert.Equal(t, "testuser", r.Username)
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}
}

func TestUserAPIGet_NoUser(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	resp, err := h.UserAPIGet(context.Background(), api.UserAPIGetRequestObject{})
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestUserAPIDelete_Success(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

	resp, err := h.AuthAPIRegistered(context.Background(), api.AuthAPIRegisteredRequestObject{})
	assert.Error(t, err)
	assert.Equal(t, errShouldntReach, err)

	switch resp.(type) {
	case api.AuthAPIRegistereddefaultJSONResponse:
		// No specific assertions needed for default response
	default:
		t.Fatalf("unexpected resp type: %T", resp)
	}
}

func TestAuthAPILogout(t *testing.T) {
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
	m := &MockProcess{}
	store := &MockStore{}
	h := NewHandler(store, m, m)

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
