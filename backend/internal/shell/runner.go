// Package shell provides utilities to interact with shell.
package shell

import (
	"fmt"
	"hash/maphash"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"omar-kada/air-compose/internal/events"
)

// Executor abstracts writing content to a file
type Executor interface {
	Exec(cmd string, args ...string) ([]byte, error)
}

type cmdExecuter struct{}

// NewExecutor creates and new Writer and returns it
func NewExecutor() Executor {
	return cmdExecuter{}
}

// Run runs a shell command and returns error if any
func (cmdExecuter) Exec(cmd string, args ...string) ([]byte, error) {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return nil, fmt.Errorf("executable not found: %w", err)
	}
	fullCmd := cmd + " " + strings.Join(args, " ")
	hash := "#" + fmt.Sprintf("%06x", maphash.String(maphash.MakeSeed(), fullCmd+time.Now().GoString()))[:6]

	slog.Debug("executing " + hash + ": " + fullCmd)
	c := execCommand(path, args...)
	c.Stderr = events.NewSlogWriter(slog.LevelDebug, hash)

	out, err := c.Output()
	slog.Debug("command result "+hash+":", "out", out, "err", err)
	return out, err
}

// execCommand is a wrapper for exec.Command for testability
var execCommand = defaultExecCommand

func defaultExecCommand(cmd string, args ...string) *exec.Cmd {
	return exec.Command(cmd, args...)
}
