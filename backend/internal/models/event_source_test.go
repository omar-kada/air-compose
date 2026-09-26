package models

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigChangedEvent(t *testing.T) {
	oldCfg := Config{Environment: Environment{"OLD": "1"}}
	newCfg := Config{Environment: Environment{"NEW": "2"}}

	evt := NewConfigChangedEvent(oldCfg, newCfg)

	assert.Equal(t, EventConfigurationUpdated, evt.Type)
	assert.Empty(t, evt.Msg)

	data, ok := evt.Data.(EventDataChange[Config])
	assert.True(t, ok)
	assert.Equal(t, oldCfg, data.Old)
	assert.Equal(t, newCfg, data.New)
}

func TestNewHealthChangedEvent(t *testing.T) {
	oldHealth := ContainerUnhealthy
	newHealth := ContainerHealthy
	unhealthy := []string{"svc1", "svc2"}

	evt := NewHealthChangedEvent(oldHealth, newHealth, unhealthy)

	assert.Equal(t, EventHealthChange, evt.Type)
	assert.Equal(t, "(healthy) svc1, svc2", evt.Msg)

	data, ok := evt.Data.(EventDataChange[ContainerHealth])
	assert.True(t, ok)
	assert.Equal(t, oldHealth, data.Old)
	assert.Equal(t, newHealth, data.New)
}

func TestNewHealthChangedEvent_EmptyUnhealthy(t *testing.T) {
	evt := NewHealthChangedEvent(ContainerStarting, ContainerStarting, []string{})

	assert.Equal(t, EventHealthChange, evt.Type)
	assert.Equal(t, "(starting) ", evt.Msg)
}

func TestNewNewCommitEvent(t *testing.T) {
	patch := Patch{Title: "Add feature", Diff: "diff content", Author: "dev"}

	evt := NewNewCommitEvent(patch)

	assert.Equal(t, EventNewCommit, evt.Type)
	assert.Equal(t, "Add feature", evt.Msg)
	assert.Equal(t, patch, evt.Data)
}

func TestFromSourceEvent_WithObjectContext(t *testing.T) {
	src := SourceEvent{Type: EventNewCommit, Msg: "commit msg", Data: Patch{Title: "t"}}
	ctx := GetDeploymentContext(context.Background(), Deployment{ID: 42, Title: "deploy-1"})

	evt := FromSourceEvent(ctx, src)

	assert.Equal(t, EventNewCommit, evt.Type)
	assert.Equal(t, "commit msg", evt.Msg)
	assert.Equal(t, uint64(42), evt.ObjectID)
	assert.Equal(t, "deploy-1", evt.ObjectName)
	assert.Equal(t, src.Data, evt.Data)
}

func TestFromSourceEvent_EmptyContext(t *testing.T) {
	src := SourceEvent{Type: EventNewCommit, Msg: "msg"}
	ctx := context.Background()

	evt := FromSourceEvent(ctx, src)

	assert.Equal(t, uint64(0), evt.ObjectID)
	assert.Empty(t, evt.ObjectName)
	assert.Equal(t, src.Type, evt.Type)
}

func TestGetObjectFromContext_WithValues(t *testing.T) {
	ctx := GetDeploymentContext(context.Background(), Deployment{ID: 99, Title: "my-deploy"})

	id, name := GetObjectFromContext(ctx)
	assert.Equal(t, uint64(99), id)
	assert.Equal(t, "my-deploy", name)
}

func TestGetObjectFromContext_EmptyContext(t *testing.T) {
	ctx := context.Background()
	id, name := GetObjectFromContext(ctx)
	assert.Equal(t, uint64(0), id)
	assert.Empty(t, name)
}

func TestGetDeploymentContext(t *testing.T) {
	ctx := GetDeploymentContext(context.Background(), Deployment{ID: 7, Title: "test-deploy"})
	assert.NotNil(t, ctx)

	id, name := GetObjectFromContext(ctx)
	assert.Equal(t, uint64(7), id)
	assert.Equal(t, "test-deploy", name)
}
