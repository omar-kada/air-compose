package models

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

// DeploymentParams groups parameters related to the deployment process
type DeploymentParams struct {
	WorkingDir   string
	ServicesDir  string
	AddWritePerm string
}

// GetAddWritePerm returns the permissions to add based on the AddWritePerm boolean
func (p DeploymentParams) GetAddWritePerm() os.FileMode {
	addPermBool, err := strconv.ParseBool(p.AddWritePerm)
	if err != nil {
		slog.Debug(fmt.Sprintf("invalid param AddWritePerm = %v", p.AddWritePerm))
	}
	if addPermBool {
		return os.FileMode(0o666)
	}
	return os.FileMode(0o000)
}

// GetRepoDir returns the path of the repo directory
func (p DeploymentParams) GetRepoDir() string {
	return filepath.Join(p.WorkingDir, "repo")
}

// GetDBDir returns the path of the database directory
func (p DeploymentParams) GetDBDir() string {
	return filepath.Join(p.WorkingDir, "db")
}

// ServerParams groups parameters related to the API server
type ServerParams struct {
	Port int
}
