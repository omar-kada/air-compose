package docker

import (
	"context"
	"fmt"
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
	GetChannel() <-chan models.ContainerHealth
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
		slog.Error("error getting stacks current state", "err", err)
		hc.eventPublisher.Publish(context.Background(), models.SourceEvent{
			Type: models.EventError,
			Msg:  "error getting stacks current state",
		})
		return
	}
	slog.Debug("[HEALTH CHECK] result", "state", state.GetGlobalHealth())
	hc.setCurrentState(state)
}

func (hc *healthChecker) setCurrentState(newState models.StacksState) {
	var globalHealth = newState.GetGlobalHealth()
	oldGlobalHealth := hc.currentStacksState.GetGlobalHealth()

	select {
	case hc.healthCheckChan <- globalHealth: // try to send
	default:
	}
	if oldGlobalHealth != globalHealth {
		if globalHealth == models.ContainerUnhealthy {
			hc.eventPublisher.Publish(context.Background(), models.SourceEvent{
				Type: models.EventStacksUnhealthy,
				Msg:  fmt.Sprintf("services : %v", strings.Join(newState.GetUnhealthyServices(), ", ")),
			})
		} else if globalHealth == models.ContainerHealthy && hc.currentStacksState != nil {
			slog.Debug("[HEALTH CHECK] sending healthy event", "oldHealth", oldGlobalHealth, "newHealth", globalHealth)
			hc.eventPublisher.Publish(context.Background(), models.SourceEvent{
				Type: models.EventStacksHealthy,
				Msg:  "",
			})
		}
	}
	hc.currentStacksState = newState
}
