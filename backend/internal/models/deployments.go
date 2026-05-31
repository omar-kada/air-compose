package models

import (
	"time"
)

// DeploymentStatus defines model for Deployment.Status.
type DeploymentStatus string

// Defines values for DeploymentStatus.
const (
	DeploymentStatusError   DeploymentStatus = "error"
	DeploymentStatusPlanned DeploymentStatus = "planned"
	DeploymentStatusRunning DeploymentStatus = "running"
	DeploymentStatusSuccess DeploymentStatus = "success"
)

// Deployment defines a deployment
type Deployment struct {
	ID      uint64 `gorm:"primaryKey;autoIncrement:true"`
	Author  string
	Diff    string
	Status  DeploymentStatus `gorm:"type:varchar(32)"`
	Time    time.Time        `gorm:"autoCreateTime"`
	EndTime time.Time
	Title   string
	Files   []FileDiff `gorm:"foreignKey:DeploymentID;constraint:OnDelete:CASCADE;"`
	Commit  string
	Repo    string
	Branch  string
}

// Compare compares two deployments by their ID.
func (d Deployment) Compare(other Deployment) int {
	if d.ID < other.ID {
		return -1
	} else if d.ID > other.ID {
		return 1
	}
	return 0
}

// FileDiff defines model for FileDiff.
type FileDiff struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement:true"`
	Diff         string
	NewFile      string
	OldFile      string
	DeploymentID uint64 `gorm:"index"`
}

// Patch represent the difference between two commits
type Patch struct {
	Diff       string
	Title      string
	Files      []FileDiff
	Author     string
	CommitHash string
}
