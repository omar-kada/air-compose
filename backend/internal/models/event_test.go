package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventToText(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected string
	}{
		{"Misc", Event{Type: EventMisc}, "Miscellaneous event"},
		{"Error", Event{Type: EventError}, "Error occurred"},
		{"DeploymentStarted", Event{Type: EventDeploymentStarted}, "Deployment started"},
		{"DeploymentSuccess", Event{Type: EventDeploymentSuccess}, "Deployment succeeded"},
		{"DeploymentError", Event{Type: EventDeploymentError}, "Deployment failed"},
		{"HealthChange", Event{Type: EventHealthChange}, "Health status"},
		{"NewCommit", Event{Type: EventNewCommit}, "New commit"},
		{"ConfigurationUpdated", Event{Type: EventConfigurationUpdated}, "Configuration updated"},
		{"PasswordUpdated", Event{Type: EventPasswordUpdated}, "Password updated"},
		{"SessionReused", Event{Type: EventSessionReused}, "Session reused"},
		{"Unknown", Event{Type: "UNKNOWN"}, "Unknown event type: UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.event.ToText())
		})
	}
}

func TestEventToEmoji(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected string
	}{
		{"Misc", Event{Type: EventMisc}, "⚪"},
		{"Error", Event{Type: EventError}, "❌"},
		{"DeploymentStarted", Event{Type: EventDeploymentStarted}, "⚙️"},
		{"DeploymentSuccess", Event{Type: EventDeploymentSuccess}, "✅"},
		{"DeploymentError", Event{Type: EventDeploymentError}, "🔴"},
		{"HealthChange", Event{
			Type: EventHealthChange,
			Data: EventDataChange[ContainerHealth]{Old: ContainerUnhealthy, New: ContainerHealthy},
		}, "✅"},
		{"NewCommit", Event{Type: EventNewCommit}, "📦"},
		{"ConfigurationUpdated", Event{Type: EventConfigurationUpdated}, "🔄"},
		{"PasswordUpdated", Event{Type: EventPasswordUpdated}, "🔑"},
		{"SessionReused", Event{Type: EventSessionReused}, "�"},
		{"Unknown", Event{Type: "UNKNOWN"}, "❓"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.event.ToEmoji())
		})
	}
}

func TestEventHealthEmoji(t *testing.T) {
	tests := []struct {
		name     string
		health   ContainerHealth
		expected string
	}{
		{"Starting", ContainerStarting, "▶️"},
		{"NoHealth", ContainerNoHealth, "⚠️"},
		{"Healthy", ContainerHealthy, "✅"},
		{"Unhealthy", ContainerUnhealthy, "🔴"},
		{"Unknown", ContainerHealth("unknown"), "❓"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, (&Event{}).healthEmoji(tt.health))
		})
	}
}

func TestEventToEmojiHealthChangeAllStates(t *testing.T) {
	tests := []struct {
		name     string
		health   ContainerHealth
		expected string
	}{
		{"Starting", ContainerStarting, "▶️"},
		{"NoHealth", ContainerNoHealth, "⚠️"},
		{"Healthy", ContainerHealthy, "✅"},
		{"Unhealthy", ContainerUnhealthy, "🔴"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Event{
				Type: EventHealthChange,
				Data: EventDataChange[ContainerHealth]{New: tt.health},
			}
			assert.Equal(t, tt.expected, e.ToEmoji())
		})
	}
}
