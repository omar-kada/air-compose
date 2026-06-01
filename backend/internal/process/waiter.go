package process

import (
	"errors"
	"time"
)

// ErrTimeout is returned when a timeout is reached.
var ErrTimeout = errors.New("timeout reached")

// WaitFor repeatedly calls fn at the specified refresh interval until it returns true,
// or until the timeout is reached. Returns true if fn returns true, or false with
// ErrTimeout if the timeout is reached.
func WaitFor(fn func() bool, refreshTime time.Duration, timeout time.Duration) (bool, error) {
	ticker := time.NewTicker(refreshTime)
	defer ticker.Stop()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			if fn() {
				return true, nil
			}
		case <-timer.C:
			return false, ErrTimeout
		}
	}
}

// func (s *service) isDeploymentEnded(cfg models.Config) bool {
// 	state, err := s.containersInspector.GetCurrentStacks(cfg.GetEnabledServices())
// 	if err != nil {
// 		return false
// 	}
// 	return state.GetGlobalHealth() == models.ContainerHealthy || state.GetGlobalHealth() == models.ContainerUnhealthy
// }
