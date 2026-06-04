package process

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// CronScheduler is responsible for cron running a sheduled job with updated config
type CronScheduler interface {
	Schedule(fn func(), newCronPeriod string) (*cron.Cron, error)
	ReSchedule(newCronPeriod string) (*cron.Cron, error)
	GetNext() time.Time
}

// NewCronScheduler creates a new ConfigScheduler that ensures only one cron job runs at a time.
func NewCronScheduler() CronScheduler {
	return &AtomicScheduler{}
}

// AtomicScheduler runs only a single cron job at a time
type AtomicScheduler struct {
	fn   func()
	cron *cron.Cron
	mu   sync.Mutex
}

// Schedule stops the old cron when it exists, and runs a new cron job
func (a *AtomicScheduler) Schedule(fn func(), newCronPeriod string) (*cron.Cron, error) {
	// make sure only one sync job is running at a time
	a.mu.Lock()
	defer a.mu.Unlock()
	a.fn = fn

	if a.cron != nil {
		a.cron.Stop()
		a.cron = nil
	}

	if newCronPeriod == "1" {
		slog.Debug("running job for a single time")
		fn()
		return nil, nil
	} else if newCronPeriod != "" && newCronPeriod != "0" {

		slog.Debug("scheduling a new cron job")
		c := cron.New()
		_, err := c.AddFunc(newCronPeriod, fn)
		if err != nil {
			return nil, err
		}
		c.Start()
		a.cron = c
		return c, nil
	}

	return nil, fmt.Errorf("couldn't schedule job, no cron period is defined")
}

// ReSchedule stops the current cron job and schedules a new one with the same function.
func (a *AtomicScheduler) ReSchedule(newCronPeriod string) (*cron.Cron, error) {
	return a.Schedule(a.fn, newCronPeriod)
}

// GetNext returns the next scheduled time of the cron job.
// If no cron job is scheduled or no entries are present, it returns the zero time.
func (a *AtomicScheduler) GetNext() time.Time {
	if a.cron == nil {
		return time.Time{}
	}
	entries := a.cron.Entries()
	if len(entries) == 0 {
		return time.Time{}
	}
	return entries[0].Next
}
