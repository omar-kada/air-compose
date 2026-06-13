package git

import (
	"errors"
	"os"
	"strings"
	"testing"

	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/testutil"

	"github.com/go-git/go-git/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mockConfig = models.Config{
	Settings: models.Settings{
		Git: models.GitConfig{
			Repo:   "https://github.com/test/repo.git",
			Branch: "main",
		},
	},
}

var (
	ErrClearRepo      = errors.New("clear repo error")
	ErrCheckoutBranch = errors.New("checkout branch error")
	ErrPullBranch     = errors.New("pull branch error")
	ErrDiffWithRemote = errors.New("diff with remote error")
)

func assertBranch(t *testing.T, clonePath string, branch string) {
	t.Helper()
	r, err := git.PlainOpen(clonePath)
	require.NoError(t, err)

	ref, err := r.Head()
	require.NoError(t, err)

	assert.Equal(t, ref.Name().Short(), branch)
}

func assertFileContent(t *testing.T, filePath string, wantContent string) {
	t.Helper()
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	assert.Equal(t, wantContent, string(content))
}

func TestClearRepo(t *testing.T) {
	fetcher := NewFetcher(os.FileMode(0o000), t.TempDir()+"/clone-repo", testutil.NewConfigGetter(mockConfig))

	err := fetcher.ClearRepo()
	assert.NoError(t, err)

	// Verify the repo directory is removed
	_, err = os.Stat(t.TempDir() + "/clone-repo")
	if !os.IsNotExist(err) {
		t.Fatalf("Expected repo directory to be removed, but it still exists")
	}
}

func TestPullBranch(t *testing.T) {
	remoteRepoPath := testutil.SetupRemoteRepo(t)
	mockConfig.Settings.Git.Repo = remoteRepoPath
	clonePath := t.TempDir() + "/clone-repo"
	fetcher := NewFetcher(os.FileMode(0o000), clonePath, testutil.NewConfigGetter(mockConfig))

	testutil.AddCommitToRepo(t, remoteRepoPath, "README.md", []byte("dummy readme"))

	// Use the fetcher methods directly
	err := fetcher.PullBranch("main", "")
	assert.NoError(t, err)
	assertFileContent(t, clonePath+"/README.md", "dummy readme")
	assertBranch(t, clonePath, "main")
}

func TestDiffWithRemote(t *testing.T) {
	clonePath := t.TempDir() + "/clone-repo"
	remoteRepoPath := testutil.SetupRemoteRepo(t)
	mockConfig.Settings.Git.Repo = remoteRepoPath
	fetcher := NewFetcher(os.FileMode(0o000), clonePath, testutil.NewConfigGetter(mockConfig))

	// Initial pull to set up the repo
	err := fetcher.PullBranch("main", "")
	assert.NoError(t, err)

	// Add a commit to the remote repo
	testutil.AddCommitToRepo(t, remoteRepoPath, "NEWFILE.txt", []byte("new file content"))

	// Get the diff with remote
	patch, err := fetcher.DiffWithRemote()
	assert.NoError(t, err)

	wantDiff := strings.Join([]string{
		"+++ b/NEWFILE.txt",
		"@@ -0,0 +1 @@",
		"+new file content",
	}, "\n")

	assert.Contains(t, patch.Diff, wantDiff)
	assert.Equal(t, "Test", patch.Author)
	assert.Equal(t, "add NEWFILE.txt", patch.Title)
	assert.Equal(t, 1, len(patch.Files))
	assert.Equal(t, "NEWFILE.txt", patch.Files[0].NewFile)
}

func TestDiffWithRemote_NoChanges(t *testing.T) {
	clonePath := t.TempDir() + "/clone-repo"
	remoteRepoPath := testutil.SetupRemoteRepo(t)
	mockConfig.Settings.Git.Repo = remoteRepoPath
	fetcher := NewFetcher(os.FileMode(0o000), clonePath, testutil.NewConfigGetter(mockConfig))

	err := fetcher.PullBranch("main", "")
	assert.NoError(t, err)

	// Get the diff with remote
	patch, err := fetcher.DiffWithRemote()
	assert.NoError(t, err)
	assert.Equal(t, "", patch.Diff)
}

func TestPullBranch_NonExistentRepo(t *testing.T) {
	clonePath := t.TempDir() + "/clone-repo"

	fetcher := NewFetcher(os.FileMode(0o000), clonePath, testutil.NewConfigGetter(models.Config{
		Settings: models.Settings{
			Git: models.GitConfig{
				Repo:   "/path/does/not/exist",
				Branch: "main",
			},
		},
	}))
	err := fetcher.PullBranch("main", "")
	assert.Error(t, err)
}

func TestFetch_WithAddPermissions(t *testing.T) {
	clonePath := t.TempDir() + "/clone-repo"
	remoteRepoPath := testutil.SetupRemoteRepo(t)
	mockConfig.Settings.Git.Repo = remoteRepoPath
	fetcher := NewFetcher(os.FileMode(0o755), clonePath, testutil.NewConfigGetter(mockConfig))

	testutil.AddCommitToRepo(t, remoteRepoPath, "README.md", []byte("dummy readme"))

	err := fetcher.PullBranch("main", "")
	assert.NoError(t, err)

	// Check file permissions
	fileInfo, err := os.Stat(clonePath + "/README.md")
	assert.NoError(t, err)

	expectedPerm := os.FileMode(0o755)
	assert.Equal(t, expectedPerm, fileInfo.Mode().Perm())
}

func TestTestGitConnection(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		remoteRepoPath := testutil.SetupRemoteRepo(t)
		fetcher := NewFetcher(os.FileMode(0o000), t.TempDir(), testutil.NewConfigGetter(mockConfig))

		success, err := fetcher.TestGitConnection(remoteRepoPath, "main", "", "")
		assert.NoError(t, err)
		assert.True(t, success)
	})

	t.Run("failed connection", func(t *testing.T) {
		fetcher := NewFetcher(os.FileMode(0o000), t.TempDir(), testutil.NewConfigGetter(mockConfig))

		success, err := fetcher.TestGitConnection("invalid-repo-url", "main", "", "")
		assert.Error(t, err)
		assert.False(t, success)
	})

	t.Run("with authentication", func(t *testing.T) {
		remoteRepoPath := testutil.SetupRemoteRepo(t)
		mockConfig.Settings.Git.Username = "testuser"
		mockConfig.Settings.Git.Token = "testtoken"
		fetcher := NewFetcher(os.FileMode(0o000), t.TempDir(), testutil.NewConfigGetter(mockConfig))

		success, err := fetcher.TestGitConnection(remoteRepoPath, "main", "testuser", "testtoken")
		assert.NoError(t, err)
		assert.True(t, success)
	})
}
func TestIsRemoteSameAsConfig(t *testing.T) {
	t.Run("repo does not exist", func(t *testing.T) {
		clonePath := t.TempDir() + "/clone-repo"
		fetcher := NewFetcher(os.FileMode(0o000), clonePath, testutil.NewConfigGetter(mockConfig))

		same, err := fetcher.IsRemoteSameAsConfig()
		assert.NoError(t, err)
		assert.False(t, same)
	})

	t.Run("repo exists with matching remote", func(t *testing.T) {
		remoteRepoPath := testutil.SetupRemoteRepo(t)
		clonePath := t.TempDir() + "/clone-repo"
		mockConfig.Settings.Git.Repo = remoteRepoPath
		fetcher := NewFetcher(os.FileMode(0o000), clonePath, testutil.NewConfigGetter(mockConfig))

		// Clone the repo first
		_, err := git.PlainClone(clonePath, &git.CloneOptions{
			URL: remoteRepoPath,
		})
		require.NoError(t, err)

		same, err := fetcher.IsRemoteSameAsConfig()
		assert.NoError(t, err)
		assert.True(t, same)
	})

	t.Run("repo exists with different remote", func(t *testing.T) {
		remoteRepoPath := testutil.SetupRemoteRepo(t)
		clonePath := t.TempDir() + "/clone-repo"
		mockConfig.Settings.Git.Repo = "https://github.com/different/repo.git"
		fetcher := NewFetcher(os.FileMode(0o000), clonePath, testutil.NewConfigGetter(mockConfig))

		// Clone the repo first
		_, err := git.PlainClone(clonePath, &git.CloneOptions{
			URL: remoteRepoPath,
		})
		require.NoError(t, err)

		same, err := fetcher.IsRemoteSameAsConfig()
		assert.NoError(t, err)
		assert.False(t, same)
	})

	t.Run("repo exists with no remote URLs", func(t *testing.T) {
		clonePath := t.TempDir() + "/clone-repo"
		// Create a bare repo without any remotes
		_, err := git.PlainInit(clonePath, true)
		require.NoError(t, err)

		// Create .git folder
		err = os.MkdirAll(clonePath+"/.git", 0o700)
		require.NoError(t, err)
		mockConfig.Settings.Git.Repo = ""
		fetcher := NewFetcher(os.FileMode(0o000), clonePath, testutil.NewConfigGetter(mockConfig))

		same, err := fetcher.IsRemoteSameAsConfig()
		assert.Error(t, err)
		assert.False(t, same)
	})
}
