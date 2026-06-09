// Package shell provides utilities to interact with shell.
package shell

import (
	"fmt"
	"log/slog"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"
	"os/exec"
	"strings"
)

// Executor abstracts writing content to a file
type Executor interface {
	Exec(cmd string, args ...string) ([]byte, error)
}

type cmdExecuter struct {
	showLogs bool
}

// NewExecutor creates and new Writer and returns it
func NewExecutor() Executor {
	features := models.LoadFeatures()
	return cmdExecuter{
		showLogs: features.DisplayCmdLogs,
	}
}

// Run runs a shell command and returns error if any
func (e cmdExecuter) Exec(cmd string, args ...string) ([]byte, error) {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return nil, fmt.Errorf("executable not found: %w", err)
	}
	fullCmd := cmd + " " + strings.Join(args, " ")
	slog.Debug("[CMD] " + fullCmd)
	c := execCommand(path, args...)
	if e.showLogs {
		c.Stderr = events.NewSlogWriter(slog.LevelDebug, cmd+"(error)")
	}

	out, err := c.Output()
	if e.showLogs {
		slog.Debug("[CMD] "+cmd, "out", string(out))
	}
	return out, err
}

// execCommand is a wrapper for exec.Command for testability
var execCommand = defaultExecCommand

func defaultExecCommand(cmd string, args ...string) *exec.Cmd {
	return exec.Command(cmd, args...)
}
