package docker

import (
	"context"
	"log/slog"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"strings"
	"sync"
	"time"
)

// HealthChecker defines the interface for checking the health of Docker stacks.
type HealthChecker interface {
	ScheduleStateRefresh(ctx context.Context)
	Get() models.StacksState
}

type healthChecker struct {
	configStore    models.ConfigGetter
	inspector      Inspector
	eventPublisher events.Publisher

	healthCheckChan    chan models.ContainerHealth
	currentStacksState models.StacksState
	refreshDuration    time.Duration
	refreshMu          sync.Mutex
}

// NewHealthChecker creates a new healthChecker instance with the given dependencies.
func NewHealthChecker(configStore models.ConfigGetter, inspector Inspector, eventPublisher events.Publisher) HealthChecker {
	hc := healthChecker{
		configStore:     configStore,
		inspector:       inspector,
		eventPublisher:  eventPublisher,
		healthCheckChan: make(chan models.ContainerHealth, 1),
		refreshDuration: 20 * time.Second,
	}
	return &hc
}

func (hc *healthChecker) ScheduleStateRefresh(ctx context.Context) {
	ticker := time.NewTicker(hc.refreshDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if hc.refreshMu.TryLock() {
				hc.refreshState()
				hc.refreshMu.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (hc *healthChecker) refreshState() {
	cfg := hc.configStore.Get()
	state, err := hc.inspector.GetCurrentStacks(cfg.GetEnabledServices())
	if err != nil {
		slog.Error("error getting stacks current state", "err", err)
		hc.eventPublisher.Publish(context.Background(), models.SourceEvent{
			Type: models.EventError,
			Msg:  "error getting stacks current state",
		})
		return
	}
	hc.setCurrentState(state)
}

func (hc *healthChecker) setCurrentState(newState models.StacksState) {

	unhealthy := newState.GetUnhealthyServices()
	var globalHealth = newState.GetGlobalHealth()
	oldGlobalHealth := hc.currentStacksState.GetGlobalHealth()

	if len(unhealthy) > 0 {
		slog.Debug("[HEALTH CHECK] unhealthy services", "services", strings.Join(unhealthy, ", "))
	} else {
		slog.Debug("[HEALTH CHECK] result", "state", globalHealth)
	}

	if oldGlobalHealth != globalHealth && (globalHealth != models.ContainerHealthy || hc.currentStacksState != nil) {
		hc.eventPublisher.Publish(context.Background(),
			models.NewHealthChangedEvent(oldGlobalHealth, globalHealth, unhealthy))
	}

	hc.currentStacksState = newState
}

func (hc *healthChecker) Get() models.StacksState {
	return hc.currentStacksState
}
