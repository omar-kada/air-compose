package models

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAddWritePerm(t *testing.T) {
	tests := []struct {
		name     string
		addPerm  string
		expected os.FileMode
	}{
		{"true", "true", os.FileMode(0o666)},
		{"True", "True", os.FileMode(0o666)},
		{"1", "1", os.FileMode(0o666)},
		{"false", "false", os.FileMode(0o000)},
		{"0", "0", os.FileMode(0o000)},
		{"invalid", "invalid", os.FileMode(0o000)},
		{"empty", "", os.FileMode(0o000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := DeploymentParams{AddWritePerm: tt.addPerm}
			assert.Equal(t, tt.expected, p.GetAddWritePerm())
		})
	}
}

func TestGetRepoDir(t *testing.T) {
	p := DeploymentParams{WorkingDir: "/tmp/air-compose"}
	assert.Equal(t, filepath.Join("/tmp/air-compose", "repo"), p.GetRepoDir())
}

func TestGetDBDir(t *testing.T) {
	p := DeploymentParams{WorkingDir: "/tmp/air-compose"}
	assert.Equal(t, filepath.Join("/tmp/air-compose", "db"), p.GetDBDir())
}

func TestIsHumanLog(t *testing.T) {
	tests := []struct {
		name     string
		humanLog string
		expected bool
	}{
		{"True (uppercase)", "TRUE", true},
		{"True (mixed case)", "True", true},
		{"True (lowercase)", "true", true},
		{"False", "FALSE", false},
		{"Empty", "", false},
		{"Other", "yes", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lp := LoggerParams{HumanLog: tt.humanLog}
			assert.Equal(t, tt.expected, lp.IsHumanLog())
		})
	}
}

func TestToLevel(t *testing.T) {
	lp := LoggerParams{LogLevel: "debug"}
	assert.Equal(t, slog.LevelDebug, lp.ToLevel())
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{"Debug", "debug", slog.LevelDebug},
		{"DEBUG (upper)", "DEBUG", slog.LevelDebug},
		{"Info", "info", slog.LevelInfo},
		{"Warn", "warn", slog.LevelWarn},
		{"Warning", "warning", slog.LevelWarn},
		{"Error", "error", slog.LevelError},
		{"Any", "any", slog.LevelInfo},
		{"Empty", "", slog.LevelInfo},
		{"Unknown", "unknown", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseLogLevel(tt.input))
		})
	}
}
