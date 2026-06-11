package socket

import (
	"context"
	"log/slog"
	"time"

	"omar-kada/air-compose/api"
)

// StateHandler handles state-related messages and connection events
type StateHandler struct {
	logger *slog.Logger
	sender MessageSender
}

// NewStateHandler creates a new StateHandler instance
func NewStateHandler(logger *slog.Logger, sender MessageSender) *StateHandler {
	return &StateHandler{
		logger: logger,
		sender: sender,
	}
}

// OnConnect is called when a new connection is established
func (sh *StateHandler) OnConnect(ctx context.Context) {
	sh.logger.Debug("[SOCKET] state handler connected")
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				state := api.StateMessage{
					Status: "running",
				}
				sh.sender.SendStateMessage(ctx, state)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// HandleMessage is called when a new message is received
func (*StateHandler) HandleMessage(_ context.Context, _ any) {
	// StateHandler does not handle any messages
}
