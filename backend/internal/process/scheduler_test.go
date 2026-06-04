package process

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCronScheduler(t *testing.T) {
	scheduler := NewCronScheduler()

	assert.NotNil(t, scheduler, "Expected non-nil scheduler")
}

func TestSchedule(t *testing.T) {
	scheduler := NewCronScheduler().(*AtomicScheduler)

	// Create a channel to signal when the function is called
	fnCalled := make(chan bool, 1)

	// Schedule the function
	c, err := scheduler.Schedule(func() {
		fnCalled <- true
	}, "@every 1s")

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
	scheduler := NewCronScheduler().(*AtomicScheduler)

	// Create a channel to signal when the function is called
	fnCalled := make(chan bool, 1)

	// Schedule the function
	c, err := scheduler.Schedule(func() {
		fnCalled <- true
	}, "")

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
	scheduler := NewCronScheduler().(*AtomicScheduler)
	// First function call channel
	fn1Called := make(chan bool, 1)
	// Second function call channel
	fn2Called := make(chan bool, 1)

	// Schedule first function
	c1, err := scheduler.Schedule(func() {
		fn1Called <- true
	}, "@every 1s")
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
	}, "@every 1s")
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
	scheduler := NewCronScheduler().(*AtomicScheduler)

	// Create a channel to signal when the function is called
	fnCalled := make(chan bool, 1)

	// Schedule the function
	c, err := scheduler.Schedule(func() {
		fnCalled <- true
	}, "1")

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
	scheduler := NewCronScheduler().(*AtomicScheduler)

	// GetNext should return zero time when no cron is scheduled
	next := scheduler.GetNext()

	assert.Zero(t, next, "Expected zero time when no cron is scheduled")
}

func TestGetNext_AfterScheduling(t *testing.T) {
	scheduler := NewCronScheduler().(*AtomicScheduler)

	// Schedule a function
	c, err := scheduler.Schedule(func() {}, "@every 10s")
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
	scheduler := NewCronScheduler().(*AtomicScheduler)

	// Schedule first function
	c1, err := scheduler.Schedule(func() {}, "@every 1m")
	assert.NoError(t, err)
	assert.NotNil(t, c1)

	next1 := scheduler.GetNext()
	assert.NotZero(t, next1, "Expected non-zero time after first scheduling")
	assert.True(t, next1.After(time.Now()), "Expected next time to be in the future")

	// Schedule second function (replaces first)
	c2, err := scheduler.Schedule(func() {}, "@every 1m")
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
	scheduler := NewCronScheduler().(*AtomicScheduler)

	// Schedule with immediate execution
	c, err := scheduler.Schedule(func() {}, "1")
	assert.NoError(t, err)
	assert.Nil(t, c, "Expected nil cron for immediate execution")

	// GetNext should return zero time when immediate execution is used
	next := scheduler.GetNext()
	assert.Zero(t, next, "Expected zero time for immediate execution")
}

func TestReSchedule(t *testing.T) {
	scheduler := NewCronScheduler().(*AtomicScheduler)

	// Create a channel to signal when the function is called
	fnCalled := make(chan bool, 1)

	// Schedule the function
	c, err := scheduler.Schedule(func() {
		fnCalled <- true
	}, "")
	assert.Error(t, err)
	assert.Nil(t, c, "nothing should be scheduled")

	// ReSchedule the function
	c, err = scheduler.ReSchedule("@every 1s")
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
