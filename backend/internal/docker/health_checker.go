package docker

import (
	"context"
	"fmt"
	"log/slog"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/storage"
	"strings"
	"sync"
	"time"
)

// HealthChecker defines the interface for checking the health of Docker stacks.
type HealthChecker interface {
	ScheduleStateRefresh(ctx context.Context)
}

type healthChecker struct {
	configStore storage.ConfigStore
	inspector   Inspector
	dispatcher  events.Dispatcher

	currentStacksState models.StacksState
	refreshDuration    time.Duration
	refreshMu          sync.Mutex
}

// NewHealthChecker creates a new healthChecker instance with the given dependencies.
func NewHealthChecker(configStore storage.ConfigStore, inspector Inspector, dispatcher events.Dispatcher) HealthChecker {
	hc := healthChecker{
		configStore: configStore,
		inspector:   inspector,
		dispatcher:  dispatcher,

		refreshDuration: 1 * time.Minute,
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
	cfg, err := hc.configStore.Get()
	if err != nil {
		slog.Error("error while getting configuration in health refresh", "err", err)
		return
	}
	state, err := hc.inspector.GetCurrentStacks(cfg.GetEnabledServices())
	if err != nil {
		slog.Error("error while getting stacks state", "err", err)
		return
	}
	hc.setCurrentState(state)
}

func (hc *healthChecker) setCurrentState(newState models.StacksState) {
	if hc.currentStacksState.GetGlobalHealth() != newState.GetGlobalHealth() {
		if newState.GetGlobalHealth() == models.ContainerUnhealthy {
			hc.dispatcher.Dispatch(context.Background(), models.EventStacksUnhealthy,
				fmt.Sprintf("services : %v", strings.Join(newState.GetUnhealthyServices(), ", ")))
		} else if newState.GetGlobalHealth() == models.ContainerHealthy && hc.currentStacksState != nil {
			hc.dispatcher.Dispatch(context.Background(), models.EventStacksHealthy, "")
		}
	}
	hc.currentStacksState = newState
}
