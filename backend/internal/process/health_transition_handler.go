package process

import (
	"context"
	"log/slog"
	"omar-kada/air-compose/internal/models"
	"sync"
	"time"
)

// HealthTransitionHandler processes container health transitions and handles redeployment
// when containers become unhealthy.
type HealthTransitionHandler struct {
	configStore       models.ConfigGetter
	deploymentService DeploymentService

	retries       int
	mu            sync.Mutex
	currentHealth models.ContainerHealth
	retryTimer    *time.Timer
}

// NewHealthTransitionHandler creates a new HealthTransitionHandler that processes container health transitions.
func NewHealthTransitionHandler(configStore models.ConfigGetter,
	deploymentService DeploymentService) *HealthTransitionHandler {

	healthHandler := &HealthTransitionHandler{
		configStore:       configStore,
		deploymentService: deploymentService,
	}

	return healthHandler
}

// HandleHealthCheck processes container health events and triggers redeployment when containers become unhealthy.
func (h *HealthTransitionHandler) HandleHealthCheck(_ context.Context, event models.Event) {
	if event.Type != models.EventHealthChange {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.currentHealth = event.Data.(models.EventDataChange[models.ContainerHealth]).New
	switch h.currentHealth {
	case models.ContainerUnhealthy, models.ContainerNoHealth:
		h.onUnhealthy()
	case models.ContainerHealthy:
		h.stopPendingRetry()
		if h.retries != 0 {
			slog.Debug("[HEALTH CHECK] resetting number of retries to 0")
			h.retries = 0
		}
	default:
		return
	}
}

// onUnhealthy must be called with h.mu held.
func (h *HealthTransitionHandler) onUnhealthy() {
	cfg := h.configStore.Get()

	if h.retries == cfg.Settings.Schedule.RetriesOnUnhealthy {
		slog.Error("[HEALTH CHECK] max retries on unhealthy reached", "retries", h.retries)
		h.retries = h.retries + 1
		return
	} else if h.retries > cfg.Settings.Schedule.RetriesOnUnhealthy {
		return
	}

	h.retries = h.retries + 1
	slog.Debug("[HEALTH CHECK] incrementing number of retries", "retries", h.retries)

	_, err := h.deploymentService.DoDeploy(DeploymentTriggerUnhealthyStacks, models.Patch{})
	if err != nil {
		slog.Error(err.Error())
	}

	h.scheduleRetry()
}

// scheduleRetry (re)arms the single retry timer. Any previously pending timer
// is stopped first so at most one wait/retry is ever in flight.
// Must be called with h.mu held.
func (h *HealthTransitionHandler) scheduleRetry() {
	cfg := h.configStore.Get()
	delay := time.Duration(cfg.Settings.Schedule.RetryDelay) * time.Millisecond

	slog.Debug("[HEALTH CHECK] waiting before next retry", "retries", h.retries,
		"delay(ms)", cfg.Settings.Schedule.RetryDelay)

	h.stopPendingRetry()
	h.retryTimer = time.AfterFunc(delay, h.onRetryFired)
}

// stopPendingRetry cancels any in-flight retry timer. Must be called with h.mu held.
func (h *HealthTransitionHandler) stopPendingRetry() {
	if h.retryTimer != nil {
		h.retryTimer.Stop()
		h.retryTimer = nil
	}
}

// onRetryFired runs in the timer's own goroutine when the delay elapses.
func (h *HealthTransitionHandler) onRetryFired() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.retryTimer = nil // this timer has fired, nothing to stop anymore

	switch h.currentHealth {
	case models.ContainerUnhealthy, models.ContainerNoHealth:
		h.onUnhealthy()
	case models.ContainerStarting:
		h.scheduleRetry() // wait more
	default:
		return
	}
}

// ResetRetries resets the retry counter to zero and cancels any pending retry.
func (h *HealthTransitionHandler) ResetRetries() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stopPendingRetry()
	h.retries = 0
}
