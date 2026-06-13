package events

import (
	"context"
	"log/slog"
	"omar-kada/air-compose/internal/models"
	"sync"
)

// Publisher publishes events to the bus.
type Publisher interface {
	Publish(ctx context.Context, event models.SourceEvent)
}

// Handler processes a single event. Implementations must be safe for concurrent use.
type Handler interface {
	HandleEvent(ctx context.Context, event models.Event)
}

// HandlerFunc is a functional adapter for Handler, similar to http.HandlerFunc.
type HandlerFunc func(ctx context.Context, event models.Event)

// HandleEvent implements the Handler interface by calling the underlying function.
func (f HandlerFunc) HandleEvent(ctx context.Context, event models.Event) {
	f(ctx, event)
}

// Bus is a fan-out event bus. Publish is non-blocking (up to buffer capacity).
// All registered handlers receive every event concurrently.
type Bus struct {
	ch       chan models.Event
	handlers []Handler
	wg       sync.WaitGroup
}

// NewBus creates a new event bus with the given buffer size and config store.
func NewBus(bufferSize int) *Bus {
	return &Bus{
		ch: make(chan models.Event, bufferSize),
	}
}

// Register adds a handler. Must be called before Run.
func (b *Bus) Register(h ...Handler) {
	b.handlers = append(b.handlers, h...)
}

// Publish enqueues an event. Non-blocking: drops the event if the buffer is full.
// Use PublishWait if you need backpressure.
func (b *Bus) Publish(ctx context.Context, srcEvent models.SourceEvent) {
	event := models.FromSourceEvent(ctx, srcEvent)

	select {
	case b.ch <- event:
	default:
		slog.Warn("event bus buffer full, dropping event", "event", event.Type)
	}
}

// PublishWait enqueues an event, blocking until space is available or ctx is cancelled.
func (b *Bus) PublishWait(ctx context.Context, srcEvent models.SourceEvent) error {
	event := models.FromSourceEvent(ctx, srcEvent)
	select {
	case b.ch <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Run starts the dispatch loop. It exits when ctx is cancelled.
// Call wg.Wait (or just defer b.Stop) after cancelling ctx to drain in-flight handlers.
func (b *Bus) Run(ctx context.Context) {
	for {
		select {
		case event, ok := <-b.ch:
			if !ok {
				return
			}
			b.dispatch(ctx, event)
		case <-ctx.Done():
			// Drain remaining buffered events before exiting.
			for {
				select {
				case event := <-b.ch:
					b.dispatch(ctx, event)
				default:
					b.wg.Wait()
					return
				}
			}
		}
	}
}

func (b *Bus) dispatch(ctx context.Context, event models.Event) {
	for _, h := range b.handlers {
		h := h // capture for goroutine
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					slog.Error("event handler panicked",
						"event", event.Type,
						"handler", handlerName(h),
						"err", r,
					)
				}
			}()
			h.HandleEvent(ctx, event)
		}()
	}
}

func handlerName(h Handler) string {
	if named, ok := h.(interface{ Name() string }); ok {
		return named.Name()
	}
	return "unknown"
}
