package models

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SourceEvent represents an event that components can publish.
type SourceEvent struct {
	Type EventType
	Msg  string
	Data any
}

// EventDataChange represents the data for a change event.
type EventDataChange[T any] struct {
	Old T
	New T
}

// NewConfigChangedEvent creates a new configuration change event with the old and new configurations.
func NewConfigChangedEvent(oldConfig, newConfig Config) SourceEvent {
	return SourceEvent{
		Type: EventConfigurationUpdated,
		Data: EventDataChange[Config]{
			Old: oldConfig,
			New: newConfig,
		},
	}
}

// NewHealthChangedEvent creates a new health change event with the old and new health statuses.
func NewHealthChangedEvent(oldHealth, newHealth ContainerHealth, notHealthy []string) SourceEvent {
	return SourceEvent{
		Type: EventHealthChange,
		Msg:  fmt.Sprintf("(%v) %v", newHealth, strings.Join(notHealthy, ", ")),
		Data: EventDataChange[ContainerHealth]{
			Old: oldHealth,
			New: newHealth,
		},
	}
}

// NewNewCommitEvent creates a new commit event with the patch information.
func NewNewCommitEvent(patch Patch) SourceEvent {
	return SourceEvent{
		Type: EventNewCommit,
		Msg:  patch.Title,
		Data: patch,
	}
}

// FromSourceEvent creates a new Event from a SourceEvent, extracting object ID and name from the context.
func FromSourceEvent(ctx context.Context, srcEvent SourceEvent) Event {
	objectID, objectName := GetObjectFromContext(ctx)

	return Event{
		Type:       srcEvent.Type,
		Msg:        srcEvent.Msg,
		Time:       time.Now(),
		ObjectID:   objectID,
		ObjectName: objectName,
		Data:       srcEvent.Data,
	}
}

// objectIDCtxKey represent a contextkey for objectID
const objectIDCtxKey ContextKey = "OBJECT_ID"

// objectNameCtxKey represents a context key for object name
const objectNameCtxKey ContextKey = "OBJECT_NAME"

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
func GetDeploymentContext(ctx context.Context, deployment Deployment) context.Context {
	ctx = context.WithValue(ctx, objectIDCtxKey, deployment.ID)
	ctx = context.WithValue(ctx, objectNameCtxKey, deployment.Title)
	return ctx
}
