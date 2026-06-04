package models

import "time"

// EventType represents the type of event
type EventType string

const (
	// EventMisc is for events that doesn't fall in one of these categories (eg: Debug events)
	EventMisc EventType = "MISC"

	// EventError indicates that an error has occurred
	EventError EventType = "ERROR"

	// EventDeploymentStarted indicates that a deployment has started
	EventDeploymentStarted EventType = "DEPLOYMENT_STARTED"

	// EventDeploymentSuccess indicates that a deployment has succeeded
	EventDeploymentSuccess EventType = "DEPLOYMENT_SUCCESS"

	// EventDeploymentError indicates that a deployment has failed
	EventDeploymentError EventType = "DEPLOYMENT_ERROR"

	// EventStacksUnhealthy indicates that some stacks are unhealthy
	EventStacksUnhealthy EventType = "STACKS_UNHEALTHY"

	// EventStacksHealthy indicates that all stacks are back to being healthy
	EventStacksHealthy EventType = "STACKS_HEALTHY"

	// EventNewCommit indicates that the repo has been updated
	EventNewCommit EventType = "NEW_COMMIT"

	// EventConfigurationUpdated indicates that a configuration has been updated
	EventConfigurationUpdated EventType = "CONFIGURATION_UPDATED"

	// EventPasswordUpdated indicates that a password has been updated
	EventPasswordUpdated EventType = "PASSWORD_UPDATED"

	// EventSessionReused indicates that a refresh token has been reused
	EventSessionReused EventType = "SESSION_REUSED"
)

// ToText returns a human-readable string representation of the event type,
func (e EventType) ToText() string {
	switch e {
	case EventMisc:
		return "Miscellaneous event"
	case EventError:
		return "Error occurred"
	case EventDeploymentStarted:
		return "Deployment started"
	case EventDeploymentSuccess:
		return "Deployment succeeded"
	case EventDeploymentError:
		return "Deployment failed"
	case EventStacksUnhealthy:
		return "Some stacks are not healthy"
	case EventStacksHealthy:
		return "All stacks are healthy"
	case EventNewCommit:
		return "New commit"
	case EventConfigurationUpdated:
		return "Configuration updated"
	case EventPasswordUpdated:
		return "Password updated"
	case EventSessionReused:
		return "Session reused"
	default:
		return "Unknown event type: " + string(e)
	}
}

// ToEmoji returns the emoji representation of the event type
func (e EventType) ToEmoji() string {
	switch e {
	case EventMisc:
		return "⚪"
	case EventError:
		return "❌"
	case EventDeploymentStarted:
		return "🚀"
	case EventDeploymentSuccess, EventStacksHealthy:
		return "✅"
	case EventDeploymentError:
		return "🔴"
	case EventStacksUnhealthy:
		return "🔴"
	case EventNewCommit:
		return "📦"
	case EventConfigurationUpdated:
		return "🔄"
	case EventPasswordUpdated:
		return "🔑"
	case EventSessionReused:
		return "🔐"
	default:
		return "❓"
	}
}

// Event represent an event inside the deployment process
type Event struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement:true"`
	Type           EventType
	Msg            string
	Time           time.Time `gorm:"autoCreateTime"`
	ObjectID       uint64    `gorm:"index"`
	ObjectName     string
	IsNotification bool
}
