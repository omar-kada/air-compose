package events

import (
	"context"
	"log/slog"
)

// Multiplex creates n channels and forwards messages from src to all of them.
func Multiplex[T any](ctx context.Context, src <-chan T, n int) []<-chan T {
	channels := make([]chan T, n)
	out := make([]<-chan T, n)
	for i := range channels {
		channels[i] = make(chan T, 4)
		out[i] = channels[i]
	}

	go func() {
		defer func() {
			for _, ch := range channels {
				close(ch)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-src:
				if !ok {
					return
				}
				for index, ch := range channels {
					select {
					case ch <- data:
					default: // drop if slow
						slog.Debug("channel is slower than source channel", "channel N°", index)
					}
				}
			}
		}
	}()

	return out
}
