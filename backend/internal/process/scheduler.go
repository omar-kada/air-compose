package process

import (
	"fmt"
	"log/slog"
	"omar-kada/air-compose/internal/config"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// ConfigScheduler is responsible for cron running a sheduled job with updated config
type ConfigScheduler interface {
	Schedule(fn func()) (*cron.Cron, error)
	ReSchedule() (*cron.Cron, error)
	GetNext() time.Time
}

// NewConfigScheduler creates a new ConfigScheduler that ensures only one cron job runs at a time.
func NewConfigScheduler(configStore config.Store) ConfigScheduler {
	return &AtomicConfigScheduler{
		configStore: configStore,
	}
}

// AtomicConfigScheduler runs only a single cron job at a time
type AtomicConfigScheduler struct {
	configStore config.Store
	fn          func()
	cron        *cron.Cron
	mu          sync.Mutex
}

// Schedule stops the old cron when it exists, and runs a new cron job
func (a *AtomicConfigScheduler) Schedule(fn func()) (*cron.Cron, error) {
	// make sure only one sync job is running at a time
	a.mu.Lock()
	defer a.mu.Unlock()
	a.fn = fn

	if a.cron != nil {
		a.cron.Stop()
		a.cron = nil
	}

	cfg := a.configStore.Get()
	newCronPeriod := cfg.Settings.Schedule.Cron
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
func (a *AtomicConfigScheduler) ReSchedule() (*cron.Cron, error) {
	return a.Schedule(a.fn)
}

// GetNext returns the next scheduled time of the cron job.
// If no cron job is scheduled or no entries are present, it returns the zero time.
func (a *AtomicConfigScheduler) GetNext() time.Time {
	if a.cron == nil {
		return time.Time{}
	}
	entries := a.cron.Entries()
	if len(entries) == 0 {
		return time.Time{}
	}
	return entries[0].Next
}
