package process

import (
	"errors"
	"testing"
	"time"

	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/testutil/mocks"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
	mocks.Fetcher
	mocks.EventPublisher
	config.Store
	DeploymentService
	CronScheduler
}

func (m *Mock) Get() models.Config {
	args := m.Called()
	return args.Get(0).(models.Config)
}

func (m *Mock) DoDeploy(trigger DeploymentTrigger, patch models.Patch) (models.Deployment, error) {
	args := m.Called(trigger, patch)
	return args.Get(0).(models.Deployment), args.Error(1)
}

func (m *Mock) Schedule(fn func(), spec string) (*cron.Cron, error) {
	args := m.Called(fn, spec)
	return args.Get(0).(*cron.Cron), args.Error(1)
}

func (m *Mock) GetNext() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func TestDeployOnChange_Success(t *testing.T) {
	m := &Mock{}
	w := NewRepoWatcher(m, m, m, m, m).(*watcher)

	patch := models.Patch{
		Diff:  "diff",
		Title: "title",
	}
	m.Fetcher.On("IsRemoteSameAsConfig").Return(true, nil)
	m.Fetcher.On("DiffWithRemote").Return(patch, nil)
	m.EventPublisher.On("Publish", mock.Anything, models.NewNewCommitEvent(patch)).Return()
	m.On("DoDeploy", DeploymentTriggerRepoUpdated, patch).Return(models.Deployment{}, nil)

	w.DeployOnChange()

	m.AssertExpectations(t)
}
func TestDeployOnChange_NoRepo(t *testing.T) {
	m := &Mock{}
	w := NewRepoWatcher(m, m, m, m, m).(*watcher)

	patch := models.Patch{
		Diff:  "diff",
		Title: "title",
	}
	m.Fetcher.On("IsRemoteSameAsConfig").Return(false, nil)
	m.Fetcher.On("DiffWithRemote").Return(patch, nil)
	m.EventPublisher.On("Publish", mock.Anything, models.NewNewCommitEvent(patch)).Return()
	m.On("DoDeploy", DeploymentTriggerRepoUpdated, patch).Return(models.Deployment{}, nil)

	w.DeployOnChange()

	m.AssertExpectations(t)
}

func TestDeployOnChange_NoChanges(t *testing.T) {
	m := &Mock{}
	w := NewRepoWatcher(m, m, m, m, m).(*watcher)

	patch := models.Patch{
		Diff: "",
	}
	m.Fetcher.On("IsRemoteSameAsConfig").Return(true, nil)
	m.Fetcher.On("DiffWithRemote").Return(patch, nil)

	w.DeployOnChange()

	m.AssertExpectations(t)
}

func TestDeployOnChange_DiffError(t *testing.T) {
	m := &Mock{}
	w := NewRepoWatcher(m, m, m, m, m).(*watcher)

	errDiff := errors.New("diff error")
	m.Fetcher.On("IsRemoteSameAsConfig").Return(true, nil)
	m.Fetcher.On("DiffWithRemote").Return(models.Patch{}, errDiff)

	w.DeployOnChange()

	m.AssertExpectations(t)
}

func TestDeployOnChange_DeployError(t *testing.T) {
	m := &Mock{}
	w := NewRepoWatcher(m, m, m, m, m).(*watcher)

	patch := models.Patch{
		Diff:  "diff",
		Title: "title",
	}
	m.Fetcher.On("IsRemoteSameAsConfig").Return(true, nil)
	m.Fetcher.On("DiffWithRemote").Return(patch, nil)
	m.EventPublisher.On("Publish", mock.Anything, models.NewNewCommitEvent(patch)).Return()
	errDeploy := errors.New("deploy error")
	m.On("DoDeploy", DeploymentTriggerRepoUpdated, patch).Return(models.Deployment{}, errDeploy)

	w.DeployOnChange()

	m.AssertExpectations(t)
}

func TestSchedule_Success(t *testing.T) {
	m := &Mock{}
	w := NewRepoWatcher(m, m, m, m, m).(*watcher)

	cronJob := &cron.Cron{}
	m.On("Schedule", mock.Anything, "0 * * * *").Return(cronJob, nil)
	m.On("Get").Return(models.Config{
		Settings: models.Settings{
			Schedule: models.ScheduleConfig{
				Cron: "0 * * * *",
			},
		},
	})

	_, err := w.Schedule()
	assert.NoError(t, err)

	m.AssertExpectations(t)
}

func TestSchedule_Error(t *testing.T) {
	m := &Mock{}
	w := NewRepoWatcher(m, m, m, m, m).(*watcher)

	errSchedule := errors.New("schedule error")
	m.On("Schedule", mock.Anything, "0 * * * *").Return(&cron.Cron{}, errSchedule)
	m.On("Get").Return(models.Config{
		Settings: models.Settings{
			Schedule: models.ScheduleConfig{
				Cron: "0 * * * *",
			},
		},
	})
	m.EventPublisher.On("Publish", mock.Anything, models.SourceEvent{
		Type: models.EventError,
		Msg:  "failed to schedule repo polling: schedule error",
	}).Return()

	_, err := w.Schedule()
	assert.Error(t, err)
	assert.Equal(t, errSchedule, err)

	m.AssertExpectations(t)
}

func TestGetNext_Success(t *testing.T) {
	m := &Mock{}
	w := NewRepoWatcher(m, m, m, m, m).(*watcher)

	next := time.Now().Add(1 * time.Hour)
	m.On("GetNext").Return(next)

	assert.Equal(t, next, w.GetNext())

	m.AssertExpectations(t)
}

func TestGetNext_ZeroTime(t *testing.T) {
	m := &Mock{}
	w := NewRepoWatcher(m, m, m, m, m).(*watcher)

	m.On("GetNext").Return(time.Time{})

	assert.Equal(t, time.Time{}, w.GetNext())

	m.AssertExpectations(t)
}
