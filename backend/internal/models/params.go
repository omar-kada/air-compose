package models

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DeploymentParams groups parameters related to the deployment process
type DeploymentParams struct {
	WorkingDir      string
	ServicesDir     string
	ConfigFile      string
	AddWritePerm    string
	AirComposeImage string
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
	Port     int
	FrontDir string
}

// LoggerParams groups parameters related to logging configuration
type LoggerParams struct {
	LogLevel          string
	LogFile           string
	HumanLog          string
	MaxLogFileSize    int
	MaxLogHistorySize int
}

// IsHumanLog returns true if the HumanLog parameter is set to "TRUE" (case-insensitive)
func (lp LoggerParams) IsHumanLog() bool {
	return strings.EqualFold("TRUE", lp.HumanLog)
}

// ToLevel converts the LogLevel string to a slog.Level
func (lp LoggerParams) ToLevel() slog.Level {
	return ParseLogLevel(lp.LogLevel)
}

// ParseLogLevel converts a string to a slog.Level
func ParseLogLevel(value string) slog.Level {
	switch strings.ToUpper(value) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	case "ANY":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}
