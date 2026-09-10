package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"omar-kada/air-compose/testutil/mocks"

	"github.com/stretchr/testify/assert"
)

func TestRedeployCommand_Success(t *testing.T) {
	baseDir := t.TempDir()
	mocker := &mocks.Executor{}
	cmd := NewRedeployCommand(mocker)

	workingDir := filepath.Join(baseDir, "work")
	servicesDir := filepath.Join(baseDir, "services")
	os.MkdirAll(workingDir, 0o750)
	os.MkdirAll(servicesDir, 0o750)

	mocker.On(
		"Exec", "docker",
		[]string{"compose", "--project-directory", filepath.Join(servicesDir, "air-compose"), "up", "-d"},
	).Return([]byte{}, nil)

	go func() {
		cmd.SetArgs([]string{
			"-d", workingDir,
			"-s", servicesDir,
		})
		cmd.Execute()
	}()

	// Wait for the command to complete
	time.Sleep(1 * time.Second)

	// Verify the command executed successfully
	mocker.AssertExpectations(t)
}

func TestRedeployCommand_Failure(t *testing.T) {
	baseDir := t.TempDir()
	mocker := &mocks.Executor{}
	cmd := NewRedeployCommand(mocker)

	workingDir := filepath.Join(baseDir, "work")
	servicesDir := filepath.Join(baseDir, "services")
	os.MkdirAll(workingDir, 0o750)
	os.MkdirAll(servicesDir, 0o750)

	mocker.On(
		"Exec", "docker",
		[]string{"compose", "--project-directory", "air-compose", "up", "-d"},
	).Return([]byte{}, errors.New("mock error"))

	// Create a channel to capture the command's exit status
	done := make(chan error, 1)

	go func() {
		err := cmd.Execute()
		done <- err
	}()

	select {

	case cmdErr := <-done:
		assert.NotNil(t, cmdErr, "mock error")
	case <-time.After(1 * time.Second):
		assert.Fail(t, "timeout while waiting for command error")
	}
}
