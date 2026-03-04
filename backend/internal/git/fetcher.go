// Package git provides functionality to operate on Git repositories.
package git

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/models"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	gitObject "github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
)

// NoErrAlreadyUpToDate is returned when the repository is already up to date.
var NoErrAlreadyUpToDate = git.NoErrAlreadyUpToDate

// Patch represent the difference between two commits
type Patch struct {
	Diff       string
	Title      string
	Files      []models.FileDiff
	Author     string
	CommitHash string
}

// Fetcher is responsible for syncing files from repo
type Fetcher interface {
	ClearRepo() error
	CheckoutBranch(branch string) error
	PullBranch(branch string, commitSHA string) error
	DiffWithRemote() (Patch, error)
}

// Syncer is responsible for syncing files from repo
type fetcher struct {
	parser         PatchParser
	addPermissions os.FileMode
	repoDir        string
	cfgStore       storage.ConfigStore

	_cfg  models.Config
	_auth *http.BasicAuth
}

// NewFetcher creates a new Syncer and returns it
func NewFetcher(addPermissions os.FileMode, repoDir string, cfgStore storage.ConfigStore) Fetcher {
	return &fetcher{
		parser:         NewPatchParser(),
		addPermissions: addPermissions,
		repoDir:        repoDir,
		cfgStore:       cfgStore,
	}
}

// setConfig sets the configuration for the fetcher
func (f *fetcher) setConfig() error {

	cfg, err := f.cfgStore.Get()
	if err != nil {
		return err
	}
	f._cfg = cfg
	f._auth = &http.BasicAuth{
		Username: cfg.Settings.Username,
		Password: cfg.Settings.Token,
	}
	return nil
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

// CheckoutBranch checks out the given local branch. If the branch does not
// exist locally but exists on the remote (origin/<branch>), it will create
// the local branch from the remote HEAD and check it out.
func (f *fetcher) CheckoutBranch(branch string) error {
	err := f.setConfig()
	if err != nil {
		return err
	}
	_, err = f.openRepo(branch)
	if err != nil {
		return fmt.Errorf("error while opening repo: %w", err)
	}
	return nil
}

// PullBranch pulls changes for a local target branch from the remote branch
// (optionally resetting to a provided commit SHA).
func (f *fetcher) PullBranch(branch string, commitHash string) error {
	err := f.setConfig()
	if err != nil {
		return err
	}
	repo, err := f.openRepo(branch)
	if err != nil {
		return err
	}

	err = f.reset(repo, branch, commitHash)
	return err
}

func (f *fetcher) openRepo(branch string) (repo *git.Repository, err error) {
	if !repoExists(f.repoDir) {
		repo, err = git.PlainClone(f.repoDir, &git.CloneOptions{
			URL:           f._cfg.Settings.Repo,
			ReferenceName: plumbing.NewBranchReferenceName(f._cfg.GetBranch()),
			SingleBranch:  true,
			Progress:      events.NewSlogWriter(slog.LevelInfo),
			Auth:          f._auth,
		})
	} else {
		repo, err = git.PlainOpen(f.repoDir)
	}
	if err != nil {
		return repo, fmt.Errorf("error while opening repo : %w, %v", err, *f)
	}
	err = repo.Fetch(&git.FetchOptions{
		Auth: f._auth,
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

func (f *fetcher) DiffWithRemote() (Patch, error) {
	err := f.setConfig()
	if err != nil {
		return Patch{}, err
	}
	repo, err := f.openRepo(f._cfg.GetBranch())
	if err != nil {
		return Patch{}, err
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
		remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", f._cfg.GetBranch()), true)
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

func repoExists(path string) bool {
	_, e := os.Stat(filepath.Join(path, ".git"))
	//_, e2 := os.Stat(filepath.Join(path, "services"))
	return e == nil /*&& e2 == nil*/
}

func branchExists(repo *git.Repository, branch string) bool {
	_, branchErr := repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	return branchErr == nil
}

func (f *fetcher) getPatch(repo *git.Repository) (Patch, error) {
	// Get local HEAD commit
	localCommit, err := getLocalHeadCommit(repo)
	if err != nil {
		return Patch{}, err
	}

	// Get remote HEAD commit (example: origin/main)
	remoteCommit, err := f.getRemoteCommit(repo)
	if err != nil {
		return Patch{}, err
	}

	if remoteCommit.Hash.Equal(localCommit.Hash) {
		// return early when commits are the same
		return Patch{}, nil
	}

	// Extract trees for diff
	localTree, err := localCommit.Tree()
	if err != nil {
		return Patch{}, fmt.Errorf("error while getting local tree: %w", err)
	}

	remoteTree, err := remoteCommit.Tree()
	if err != nil {
		return Patch{}, fmt.Errorf("error while getting remote tree: %w", err)
	}

	// Compute patch (the diff)
	patch, err := localTree.Patch(remoteTree)
	if err != nil {
		return Patch{}, fmt.Errorf("error while getting patch: %w", err)
	}

	return f.parser.Parse(patch.String(), remoteCommit)
}

func (f *fetcher) getRemoteCommit(repo *git.Repository) (*gitObject.Commit, error) {
	remoteRefName := plumbing.NewRemoteReferenceName("origin", f._cfg.GetBranch())

	remoteRef, err := repo.Reference(remoteRefName, true)
	if err != nil {
		return nil, fmt.Errorf("error while getting remote reference: %w", err)
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
