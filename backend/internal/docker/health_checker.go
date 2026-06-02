package docker

import (
	"context"
	"fmt"
	"log/slog"
	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"strings"
	"sync"
	"time"
)

// HealthChecker defines the interface for checking the health of Docker stacks.
type HealthChecker interface {
	ScheduleStateRefresh(ctx context.Context)
	GetChannel() <-chan models.ContainerHealth
}

type healthChecker struct {
	configStore config.Store
	inspector   Inspector
	dispatcher  events.Dispatcher

	healthCheckChan    chan models.ContainerHealth
	currentStacksState models.StacksState
	refreshDuration    time.Duration
	refreshMu          sync.Mutex
}

// NewHealthChecker creates a new healthChecker instance with the given dependencies.
func NewHealthChecker(configStore config.Store, inspector Inspector, dispatcher events.Dispatcher) HealthChecker {
	hc := healthChecker{
		configStore:     configStore,
		inspector:       inspector,
		dispatcher:      dispatcher,
		healthCheckChan: make(chan models.ContainerHealth, 1),
		refreshDuration: 1 * time.Minute,
	}
	return &hc
}

func (hc *healthChecker) GetChannel() <-chan models.ContainerHealth {
	return hc.healthCheckChan
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
		slog.Error("error while getting stacks state", "err", err)
		return
	}
	slog.Debug("health check result", "state", state.GetGlobalHealth())
	hc.setCurrentState(state)
}

func (hc *healthChecker) setCurrentState(newState models.StacksState) {
	select {
	case hc.healthCheckChan <- newState.GetGlobalHealth(): // try to send
	default:
	}
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
