// Package mocks provides mock implementations for testing purposes.
package mocks

import (
	"omar-kada/air-compose/internal/models"

	"github.com/stretchr/testify/mock"
)

// Fetcher is a mock implementation of the Fetcher interface.
type Fetcher struct {
	mock.Mock
}

// ClearRepo clears the repository.
func (m *Fetcher) ClearRepo() error {
	args := m.Called()
	return args.Error(0)
}

// PullBranch pulls the specified branch and commit SHA.
func (m *Fetcher) PullBranch(branch string, commitSHA string) error {
	args := m.Called(branch, commitSHA)
	return args.Error(0)
}

// DiffWithRemote returns the patch between local and remote.
func (m *Fetcher) DiffWithRemote() (models.Patch, error) {
	args := m.Called()
	return args.Get(0).(models.Patch), args.Error(1)
}

// TestGitConnection tests the git connection with the given parameters.
func (m *Fetcher) TestGitConnection(repo, branch, username, token string) (bool, error) {
	args := m.Called(repo, branch, username, token)
	return args.Bool(0), args.Error(1)
}
