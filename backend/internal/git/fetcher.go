// Package git provides functionality to operate on Git repositories.
package git

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/models"

	"github.com/go-git/go-git/v6"
	gitConfig "github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/client"

	gitObject "github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
)

// NoErrAlreadyUpToDate is returned when the repository is already up to date.
var NoErrAlreadyUpToDate = git.NoErrAlreadyUpToDate

// Fetcher is responsible for syncing files from repo
type Fetcher interface {
	ClearRepo() error
	PullBranch(branch string, commitSHA string) error
	DiffWithRemote() (models.Patch, error)
	IsRemoteSameAsConfig() (bool, error)
	TestGitConnection(repo, branch, username, token string) (bool, error)
}

type fetcher struct {
	parser         PatchParser
	addPermissions os.FileMode
	repoDir        string
	cfgStore       models.ConfigGetter

	_auth *http.BasicAuth
}

// NewFetcher creates a new Fetcher and returns it
func NewFetcher(addPermissions os.FileMode, repoDir string, cfgStore models.ConfigGetter) Fetcher {
	return &fetcher{
		parser:         NewPatchParser(),
		addPermissions: addPermissions,
		repoDir:        repoDir,
		cfgStore:       cfgStore,
	}
}

// setConfig sets the configuration for the fetcher
func (f *fetcher) setConfig() {
	cfg := f.cfgStore.Get()
	f._auth = &http.BasicAuth{
		Username: cfg.Settings.Git.Username,
		Password: cfg.Settings.Git.Token,
	}
}

func (f *fetcher) addPerm() error {
	return filepath.Walk(f.repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		originalPerm := info.Mode().Perm()
		if err := os.Chmod(path, originalPerm|f.addPermissions); err != nil {
			return err
		}
		return nil
	})
}

// ClearRepo removes the repository directory and all its contents.
func (f *fetcher) ClearRepo() error {
	return os.RemoveAll(f.repoDir)
}

// PullBranch pulls changes for a local target branch from the remote branch
// (optionally resetting to a provided commit SHA).
func (f *fetcher) PullBranch(branch string, commitHash string) error {
	f.setConfig()
	repo, err := f.openRepo(branch)
	if err != nil {
		return err
	}

	err = f.reset(repo, branch, commitHash)
	return err
}

func (f *fetcher) openRepo(branch string) (repo *git.Repository, err error) {
	if !f.repoExists() {
		repo, err = git.PlainClone(f.repoDir, &git.CloneOptions{
			URL:           f.cfgStore.Get().Settings.Git.Repo,
			ReferenceName: plumbing.NewBranchReferenceName(f.cfgStore.Get().GetBranch()),
			SingleBranch:  true,
			Progress:      events.NewSlogWriter(slog.LevelDebug, "[GIT] clone "+branch),
			ClientOptions: []client.Option{client.WithHTTPAuth(f._auth)},
		})
	} else {
		repo, err = git.PlainOpen(f.repoDir)
	}
	if err != nil {
		return repo, fmt.Errorf("error while opening repo : %w, %v", err, *f)
	}
	err = repo.Fetch(&git.FetchOptions{
		ClientOptions: []client.Option{client.WithHTTPAuth(f._auth)},
		RefSpecs: []gitConfig.RefSpec{
			"refs/heads/*:refs/remotes/origin/*",
		},
	})

	if err != nil && err != NoErrAlreadyUpToDate {
		return repo, fmt.Errorf("error while fetching repo : %w, %v", err, *f)
	}

	if branch != "" {
		err = f.checkoutOrCreate(repo, branch)
		if err != nil {
			return repo, fmt.Errorf("error while checkout branch '%v': %w", branch, err)
		}
	}
	f.addPerm()
	return repo, nil
}

func (f *fetcher) DiffWithRemote() (models.Patch, error) {
	f.setConfig()
	repo, err := f.openRepo(f.cfgStore.Get().GetBranch())
	if err != nil {
		return models.Patch{}, err
	}

	return f.getPatch(repo)
}

func (f *fetcher) reset(repo *git.Repository, branch string, hash string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error while getting worktree: %w", err)
	}
	var targetHash plumbing.Hash
	if hash != "" {
		targetHash = plumbing.NewHash(hash)
	} else {
		remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", f.cfgStore.Get().GetBranch()), true)
		if err != nil {
			return fmt.Errorf("error while getting reference for remote branch '%v': %w", branch, err)
		}
		targetHash = remoteRef.Hash()
	}
	err = wt.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: targetHash,
	})
	if err != nil {
		return fmt.Errorf("error while resetting to commit '%v': %w", targetHash, err)
	}
	f.addPerm()
	return nil
}

func (f *fetcher) checkoutOrCreate(repo *git.Repository, branch string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error while getting worktree : %w", err)
	}
	shouldCreate := !branchExists(repo, branch)
	var targetHash plumbing.Hash
	if shouldCreate {
		remoteCommit, getCommitErr := f.getRemoteCommit(repo)
		if getCommitErr != nil {
			return fmt.Errorf("error while getting remote commit : %w", getCommitErr)
		}
		targetHash = remoteCommit.Hash
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Force:  true,
			Create: true,
			Hash:   targetHash,
		})
	} else {
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Force:  true,
		})
	}
	if err != nil {
		return fmt.Errorf("failed to checkout branch '%v' (%v, %v) : %w", branch, shouldCreate, targetHash, err)
	}

	f.addPerm()
	return nil
}

func (f *fetcher) repoExists() bool {
	_, e := git.PlainOpen(f.repoDir)
	slog.Info("repo exists ? ", "error", e)
	if e != nil {
		f.ClearRepo() // clear the repo if it's not openable
	}
	return e == nil
}

func (f *fetcher) IsRemoteSameAsConfig() (bool, error) {
	if !f.repoExists() {
		return false, nil
	}
	f.setConfig()

	repo, err := git.PlainOpen(f.repoDir)
	if err != nil {
		return false, fmt.Errorf("failed to open repository: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return false, fmt.Errorf("failed to get remote: %w", err)
	}

	remotes := remote.Config().URLs
	if len(remotes) == 0 {
		return false, errors.New("no remote URLs configured")
	}

	for _, remoteURL := range remotes {
		if remoteURL == f.cfgStore.Get().Settings.Git.Repo {
			return true, nil
		}
	}
	return false, nil
}

func branchExists(repo *git.Repository, branch string) bool {
	_, branchErr := repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	return branchErr == nil
}

func (f *fetcher) getPatch(repo *git.Repository) (models.Patch, error) {
	// Get local HEAD commit
	localCommit, err := getLocalHeadCommit(repo)
	if err != nil {
		return models.Patch{}, err
	}

	// Get remote HEAD commit (example: origin/main)
	remoteCommit, err := f.getRemoteCommit(repo)
	if err != nil {
		return models.Patch{}, err
	}

	if remoteCommit.Hash.Equal(localCommit.Hash) {
		// return early when commits are the same
		return models.Patch{}, nil
	}

	// Extract trees for diff
	localTree, err := localCommit.Tree()
	if err != nil {
		return models.Patch{}, fmt.Errorf("error while getting local tree: %w", err)
	}

	remoteTree, err := remoteCommit.Tree()
	if err != nil {
		return models.Patch{}, fmt.Errorf("error while getting remote tree: %w", err)
	}

	// Compute patch (the diff)
	patch, err := localTree.Patch(remoteTree)
	if err != nil {
		return models.Patch{}, fmt.Errorf("error while getting patch: %w", err)
	}

	return f.parser.Parse(patch.String(), remoteCommit)
}

func (f *fetcher) getRemoteCommit(repo *git.Repository) (*gitObject.Commit, error) {
	remoteRefName := plumbing.NewRemoteReferenceName("origin", f.cfgStore.Get().GetBranch())

	remoteRef, err := repo.Reference(remoteRefName, true)
	if err != nil {
		return nil, fmt.Errorf("error while getting remote reference (%v): %w", remoteRefName, err)
	}

	remoteCommit, err := repo.CommitObject(remoteRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("error while getting remote commit object: %w", err)
	}
	return remoteCommit, nil
}

func getLocalHeadCommit(repo *git.Repository) (*gitObject.Commit, error) {
	headRef, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("error while getting repo HEAD: %w", err)
	}

	localCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("error while getting commitObject : %w", err)
	}
	return localCommit, nil
}

// TestGitConnection tests the connection to a Git repository by attempting to clone it.
func (*fetcher) TestGitConnection(repo, branch, username, token string) (bool, error) {
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		return false, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)
	var auth *http.BasicAuth
	if token != "" {
		auth = &http.BasicAuth{
			Username: username,
			Password: token,
		}
	}
	refName := plumbing.NewBranchReferenceName(branch)
	if branch == "" {
		refName = plumbing.NewBranchReferenceName(models.DefaultBranch)
	}

	res, err := git.PlainClone(tempDir, &git.CloneOptions{
		URL:           repo,
		ClientOptions: []client.Option{client.WithHTTPAuth(auth)},
		ReferenceName: refName,
	})
	slog.Debug("Testing git connection", "res", res, "err", err)
	if err == nil {
		return true, nil
	}

	parts := strings.Split(err.Error(), ": ")
	if len(parts) > 1 {
		return false, errors.New(parts[len(parts)-1])
	}
	return false, err
}
