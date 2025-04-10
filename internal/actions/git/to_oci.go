package git

import (
	"context"
	"errors"
	"fmt"
	"os"

	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/errdef"

	"github.com/act3-ai/data-tool/internal/git"
	"github.com/act3-ai/data-tool/internal/git/cache"
	"github.com/act3-ai/data-tool/internal/git/cmd"
	"github.com/act3-ai/data-tool/internal/ui"
	"github.com/act3-ai/go-common/pkg/logger"
)

// ToOCI represents the git sync action.
type ToOCI struct {
	*Action

	Clean bool

	GitRemote string
	Repo      string
	RevList   []string // rev-list: Reverse Chronological Order, a git convention
}

// Run performs the ToOCI operation.
func (action *ToOCI) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	// action.Config.Repository ensures we don't use an OCI cache, which is created by
	// action.Config.GraphTarget. In the case of git, it's more reasonable to cache
	// git objects themselves rather than the higher-level bundles stored as OCI artifacts,
	// as many bundles may contain the same git objects.
	log.InfoContext(ctx, "Connecting to OCI repository", "repo", action.Repo)
	repo, err := action.Config.Repository(ctx, action.Repo)
	if err != nil {
		return fmt.Errorf("creating repository reference: %w", err)
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "ToOCI-*") // removed by toOCI.Cleanup
	if err != nil {
		return fmt.Errorf("creating temporary directory for intermediate git repository: %w", err)
	}

	fs, err := file.New(tmpDir)
	if err != nil {
		return fmt.Errorf("initializing shared filestore: %w", err)
	}
	defer fs.Close()

	cmdOpts := cmd.Options{
		GitOptions: cmd.GitOptions{
			AltGitExec: action.AltGitExec,
		},
		LFSOptions: &cmd.LFSOptions{
			WithLFS:    action.LFS,
			AltLFSExec: action.AltGitLFSExec,
			ServerURL:  action.LFSServerURL,
		},
	}

	var cacheLink cache.ObjectCacher
	if action.CacheDir != "" {
		objCache, err := cache.NewCache(ctx, action.CacheDir, tmpDir, fs, &cmdOpts)
		if err != nil {
			// continue without caching
			goto Recover
		}

		// init cache link
		link, err := objCache.NewLink(ctx, cmdOpts)
		if err != nil {
			// continue without caching
			goto Recover
		}
		cacheLink = link
	}

Recover:
	syncOpts := git.SyncOptions{
		UserAgent:         action.Config.UserAgent(),
		IntermediateDir:   tmpDir,
		IntermediateStore: fs,
		Cache:             cacheLink,
	}

	existingDesc, err := repo.Resolve(ctx, repo.Reference.Reference)
	if err != nil && !errors.Is(err, errdef.ErrNotFound) {
		return fmt.Errorf("resolving git manifest descriptor: %w", err)
	}

	toOCI, err := git.NewToOCI(ctx, repo, existingDesc, action.GitRemote, action.RevList, syncOpts, &cmdOpts)
	if err != nil {
		return fmt.Errorf("prepparing to run to-oci action: %w", err)
	}
	defer toOCI.Cleanup() //nolint

	log.InfoContext(ctx, "Starting ToOCI action")
	newDesc, err := toOCI.Run(ctx)
	if err != nil {
		return fmt.Errorf("syncing (Git) %q to (OCI) %q: %w", action.GitRemote, action.Repo, err)
	}
	rootUI.Infof("Commit Manifest digest: %s", newDesc.Digest)

	log.DebugContext(ctx, "Tagging base manifest")

	err = repo.Tag(ctx, newDesc, repo.Reference.Reference)
	if err != nil {
		return fmt.Errorf("tagging base manifest: %w", err)
	}

	if err := toOCI.Cleanup(); err != nil {
		return fmt.Errorf("cleaning up toOCI: %w", err)
	}

	if err := fs.Close(); err != nil {
		return fmt.Errorf("closing shared file store: %w", err)
	}

	return nil
}
