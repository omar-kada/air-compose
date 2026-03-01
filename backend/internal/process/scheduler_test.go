package process

import (
	"path/filepath"
	"testing"
	"time"

	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
)

// createTestConfigStore creates a ConfigStore with a temporary config file
func createTestConfigStore(t *testing.T) storage.ConfigStore {
	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, "config.yaml")
	return storage.NewConfigStore(configFilePath)
}

func TestNewConfigScheduler(t *testing.T) {
	configStore := createTestConfigStore(t)
	scheduler := NewConfigScheduler(configStore)

	assert.NotNil(t, scheduler, "Expected non-nil scheduler")
}

func TestSchedule(t *testing.T) {
	configStore := createTestConfigStore(t)
	scheduler := NewConfigScheduler(configStore).(*AtomicConfigScheduler)

	// Set up test config
	testConfig := models.Config{Settings: models.Settings{Cron: "@every 1s"}}
	err := configStore.Update(testConfig)
	assert.NoError(t, err)

	// Create a channel to signal when the function is called
	fnCalled := make(chan bool, 1)

	// Schedule the function
	c, err := scheduler.Schedule(func() {
		fnCalled <- true
	})

	// Verify the cron was created and started
	assert.NoError(t, err)
	assert.NotNil(t, c, "Expected non-nil cron")

	// Wait for the function to be called
	select {
	case <-fnCalled:
		// Function was called
	case <-time.After(2 * time.Second):
		t.Error("Function was not called within expected time")
	}

	// Stop the cron
	c.Stop()
}

func TestScheduleNoCronPeriod(t *testing.T) {
	configStore := createTestConfigStore(t)
	scheduler := NewConfigScheduler(configStore).(*AtomicConfigScheduler)

	// Set up test config with no cron period
	testConfig := models.Config{}
	err := configStore.Update(testConfig)
	assert.NoError(t, err)

	// Create a channel to signal when the function is called
	fnCalled := make(chan bool, 1)

	// Schedule the function
	c, err := scheduler.Schedule(func() {
		fnCalled <- true
	})

	// Verify the cron was not created
	assert.Error(t, err, "Expected error when no cron period is defined")
	assert.Nil(t, c, "Expected nil cron when no cron period is defined")

	// Verify the function is not called
	select {
	case <-fnCalled:
		t.Error("Function was called when no cron period was defined")
	case <-time.After(1 * time.Second):
		// Expected behavior - function not called
	}
}

func TestScheduleWhileRunning(t *testing.T) {
	configStore := createTestConfigStore(t)
	scheduler := NewConfigScheduler(configStore).(*AtomicConfigScheduler)

	// Set up test config
	testConfig := models.Config{Settings: models.Settings{Cron: "@every 1s"}}
	err := configStore.Update(testConfig)
	assert.NoError(t, err)

	// First function call channel
	fn1Called := make(chan bool, 1)
	// Second function call channel
	fn2Called := make(chan bool, 1)

	// Schedule first function
	c1, err := scheduler.Schedule(func() {
		fn1Called <- true
	})
	assert.NoError(t, err)
	assert.NotNil(t, c1, "Expected non-nil cron for first function")

	// Wait for first function to be called
	select {
	case <-fn1Called:
		// First function was called
	case <-time.After(2 * time.Second):
		t.Error("First function was not called within expected time")
	}

	// Schedule second function while first is running
	c2, err := scheduler.Schedule(func() {
		fn2Called <- true
	})
	assert.NoError(t, err)
	assert.NotNil(t, c2, "Expected non-nil cron for second function")

	// Verify both functions can be called
	select {
	case <-fn1Called:
		t.Error("First function was called again even when rescheduled")
	case <-time.After(2 * time.Second):
		t.Error("First function was not called again within expected time")
	case <-fn2Called:
	}

	// Stop both crons
	c2.Stop()
}

func TestScheduleImmediateExecution(t *testing.T) {
	configStore := createTestConfigStore(t)
	scheduler := NewConfigScheduler(configStore).(*AtomicConfigScheduler)

	// Set up test config with "1" cron period
	testConfig := models.Config{Settings: models.Settings{Cron: "1"}}

	err := configStore.Update(testConfig)
	assert.NoError(t, err)

	// Create a channel to signal when the function is called
	fnCalled := make(chan bool, 1)

	// Schedule the function
	c, err := scheduler.Schedule(func() {
		fnCalled <- true
	})

	// Verify the function was called immediately
	select {
	case <-fnCalled:
		// Function was called immediately
	case <-time.After(1 * time.Second):
		t.Error("Function was not called immediately")
	}

	// Verify no cron was created
	assert.Nil(t, c, "Expected nil cron for immediate execution")
	assert.NoError(t, err)
}

func TestGetNext_NoCronScheduled(t *testing.T) {
	configStore := createTestConfigStore(t)
	scheduler := NewConfigScheduler(configStore).(*AtomicConfigScheduler)

	// GetNext should return zero time when no cron is scheduled
	next := scheduler.GetNext()

	assert.Zero(t, next, "Expected zero time when no cron is scheduled")
}

func TestGetNext_AfterScheduling(t *testing.T) {
	configStore := createTestConfigStore(t)
	scheduler := NewConfigScheduler(configStore).(*AtomicConfigScheduler)

	// Set up test config
	testConfig := models.Config{Settings: models.Settings{Cron: "@every 10s"}}
	err := configStore.Update(testConfig)
	assert.NoError(t, err)

	// Schedule a function
	c, err := scheduler.Schedule(func() {})
	assert.NoError(t, err)
	assert.NotNil(t, c, "Expected non-nil cron")

	// GetNext should return a valid future time
	next := scheduler.GetNext()

	assert.NotZero(t, next, "Expected non-zero time after scheduling")
	assert.True(t, next.After(time.Now()), "Expected next time to be in the future")

	// Stop the cron
	c.Stop()
}

func TestGetNext_MultipleSchedules(t *testing.T) {
	configStore := createTestConfigStore(t)
	scheduler := NewConfigScheduler(configStore).(*AtomicConfigScheduler)

	// Set up test config
	testConfig := models.Config{Settings: models.Settings{Cron: "@every 1m"}}
	err := configStore.Update(testConfig)
	assert.NoError(t, err)

	// Schedule first function
	c1, err := scheduler.Schedule(func() {})
	assert.NoError(t, err)
	assert.NotNil(t, c1)

	next1 := scheduler.GetNext()
	assert.NotZero(t, next1, "Expected non-zero time after first scheduling")
	assert.True(t, next1.After(time.Now()), "Expected next time to be in the future")

	// Schedule second function (replaces first)
	c2, err := scheduler.Schedule(func() {})
	assert.NoError(t, err)
	assert.NotNil(t, c2)

	next2 := scheduler.GetNext()
	assert.NotZero(t, next2, "Expected non-zero time after second scheduling")
	assert.True(t, next2.After(time.Now()), "Expected next time to be in the future")

	// Both scheduled times should be in the same ballpark (within a few seconds)
	timeDiff := next1.Sub(next2).Abs()
	assert.True(t, timeDiff < 5*time.Second, "Expected next times to be in the same ballpark")

	// Stop the cron
	c2.Stop()
}

func TestGetNext_ImmediateExecution(t *testing.T) {
	configStore := createTestConfigStore(t)
	scheduler := NewConfigScheduler(configStore).(*AtomicConfigScheduler)

	// Set up test config with immediate execution
	testConfig := models.Config{Settings: models.Settings{Cron: "1"}}
	err := configStore.Update(testConfig)
	assert.NoError(t, err)

	// Schedule with immediate execution
	c, err := scheduler.Schedule(func() {})
	assert.NoError(t, err)
	assert.Nil(t, c, "Expected nil cron for immediate execution")

	// GetNext should return zero time when immediate execution is used
	next := scheduler.GetNext()
	assert.Zero(t, next, "Expected zero time for immediate execution")
}

func TestReSchedule(t *testing.T) {
	configStore := createTestConfigStore(t)
	scheduler := NewConfigScheduler(configStore).(*AtomicConfigScheduler)

	// Set up test config
	testConfig := models.Config{Settings: models.Settings{Cron: ""}}
	err := configStore.Update(testConfig)
	assert.NoError(t, err)

	// Create a channel to signal when the function is called
	fnCalled := make(chan bool, 1)

	// Schedule the function
	c, err := scheduler.Schedule(func() {
		fnCalled <- true
	})
	assert.Error(t, err)
	assert.Nil(t, c, "nothing should be scheduled")

	// Update the config
	testConfig = models.Config{Settings: models.Settings{Cron: "@every 1s"}}
	err = configStore.Update(testConfig)
	assert.NoError(t, err)

	// ReSchedule the function
	c, err = scheduler.ReSchedule()
	assert.NoError(t, err)
	assert.NotNil(t, c, "Expected non-nil cron after rescheduling")

	// Wait for the function to be called
	select {
	case <-fnCalled:
		// Function was called again
	case <-time.After(2 * time.Second):
		t.Error("Function was not called again within expected time")
	}

	// Stop the cron
	c.Stop()
}
