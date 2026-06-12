package socket

import (
	"context"
)

// Handler defines an interface for handling messages and connection events
type Handler interface {
	// OnConnect is called when a new connection is established
	OnConnect(ctx context.Context)

	// HandleMessage is called when a new message is received
	// msg's type is one of the defined ClientMessage types
	HandleMessage(ctx context.Context, msg any)
}
