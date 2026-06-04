package shell

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand_Success(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("echo", "success")
	}

	out, err := NewExecutor().Exec("go", "help")
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestRunCommand_NoArgs(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(_ string, _ ...string) *exec.Cmd {
		c := exec.Command("false")
		c.Stderr = nil
		return c
	}

	_, err := NewExecutor().Exec("go")
	assert.ErrorContains(t, err, "exit status 1")
}

func TestRunCommand_NotFound(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("non-existent-command")
	}

	_, err := NewExecutor().Exec("dummyCmd")
	assert.ErrorContains(t, err, "executable not found")
}

func TestRunCommand_ExecError(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return &exec.Cmd{
			Path: "invalid-path",
			Err:  errors.New("exec error"),
		}
	}

	_, err := NewExecutor().Exec("echo")
	assert.ErrorContains(t, err, "exec error")
}

func TestRunCommand_WithLogs(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("echo", "with logs")
	}

	t.Setenv("AIR_COMPOSE_DISPLAY_CMD_LOGS", "true")

	out, err := NewExecutor().Exec("go", "version")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(string(out)), "with logs")
}
