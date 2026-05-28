package docker

import (
	"context"
	"fmt"
	"log/slog"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"
	"strings"
	"time"
)

// HealthChecker defines the interface for checking the health of Docker stacks.
type HealthChecker interface {
	ScheduleStateRefresh()
}

type healthChecker struct {
	configStore storage.ConfigStore
	inspector   Inspector
	dispatcher  events.Dispatcher

	currentStacksState models.StacksState
	_refreshDuration   time.Duration
}

// NewHealthChecker creates a new healthChecker instance with the given dependencies.
func NewHealthChecker(configStore storage.ConfigStore, inspector Inspector, dispatcher events.Dispatcher) HealthChecker {
	hc := healthChecker{
		configStore: configStore,
		inspector:   inspector,
		dispatcher:  dispatcher,

		_refreshDuration: 1 * time.Minute,
	}
	return &hc
}

func (hc *healthChecker) ScheduleStateRefresh() {
	ticker := time.NewTicker(hc._refreshDuration)
	defer ticker.Stop()

	for range ticker.C {
		hc.refreshState()
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
		}
	}
	hc.currentStacksState = newState
}
