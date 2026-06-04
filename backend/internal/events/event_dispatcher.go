// Package events handles logic reltaed to events
package events

import (
	"context"
	"log/slog"
	"time"

	"omar-kada/air-compose/internal/models"
)

// objectIDCtxKey represent a contextkey for objectID
const objectIDCtxKey models.ContextKey = "OBJECT_ID"

// objectNameCtxKey represents a context key for object name
const objectNameCtxKey models.ContextKey = "OBJECT_NAME"

// Dispatcher is responsible of processing deployment events
type Dispatcher interface {
	Dispatch(ctx context.Context, eventType models.EventType, msg string)
	AddHandler(handler Handler)
}

// Handler handles events dispatched by the Dispatcher.
type Handler interface {
	HandleEvent(ctx context.Context, event models.Event)
}

type dispatcher struct {
	eventHandlers []Handler
}

// NewDefaultDispatcher creates a new event dispatcher
func NewDefaultDispatcher(eventHandlers []Handler) Dispatcher {
	return &dispatcher{
		eventHandlers: eventHandlers,
	}
}

// NewVoidDispatcher creates a new dispatcher that discards all events
// without storing or logging them.
func NewVoidDispatcher() Dispatcher {
	return &dispatcher{}
}

func (d *dispatcher) Dispatch(ctx context.Context, eventType models.EventType, msg string) {
	objectID, objectName := GetObjectFromContext(ctx)

	event := models.Event{
		Type:       eventType,
		Msg:        msg,
		Time:       time.Now(),
		ObjectID:   objectID,
		ObjectName: objectName,
	}

	for _, handler := range d.eventHandlers {
		go func(h Handler) {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("event handler panicked", "err", r)
				}
			}()
			h.HandleEvent(ctx, event)
		}(handler)
	}
}

// AddHandler adds a new event handler to the dispatcher.
func (d *dispatcher) AddHandler(handler Handler) {
	d.eventHandlers = append(d.eventHandlers, handler)
}

// GetObjectFromContext extracts object ID and name from the context.
func GetObjectFromContext(ctx context.Context) (uint64, string) {
	objectID := uint64(0)
	objectName := ""

	if ctx.Value(objectIDCtxKey) != nil {
		objectID = ctx.Value(objectIDCtxKey).(uint64)
	}
	if ctx.Value(objectNameCtxKey) != nil {
		objectName = ctx.Value(objectNameCtxKey).(string)
	}

	return objectID, objectName
}

// GetDeploymentContext adds deployment ID and title to the context.
func GetDeploymentContext(ctx context.Context, deployment models.Deployment) context.Context {
	ctx = context.WithValue(ctx, objectIDCtxKey, deployment.ID)
	ctx = context.WithValue(ctx, objectNameCtxKey, deployment.Title)
	return ctx
}
