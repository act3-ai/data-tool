package git

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/act3-ai/asce/data/tool/internal/git"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
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

	log.InfoContext(ctx, "Connecting to OCI repository", "repo", action.Repo)
	repo, err := action.Config.ConfigureRepository(ctx, action.Repo)
	if err != nil {
		return fmt.Errorf("creating repository reference: %w", err)
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "ToOCI-*") // removed by toOCI.Cleanup
	if err != nil {
		return fmt.Errorf("creating temporary directory for intermediate git repository: %w", err)
	}

	syncOpts := git.SyncOptions{
		Clean:     action.Clean,
		DTVersion: action.Version(),
		TmpDir:    tmpDir,
		CacheDir:  action.CacheDir,
	}

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

	toOCI, err := git.NewToOCI(ctx, repo, repo.Reference.Reference, action.GitRemote, action.RevList, syncOpts, &cmdOpts)
	if err != nil {
		return fmt.Errorf("prepparing to run to-oci action: %w", err)
	}
	defer toOCI.Cleanup() //nolint

	log.InfoContext(ctx, "syncing commit manifest")
	commitManDesc, err := toOCI.Run(ctx)
	if err != nil {
		return fmt.Errorf("syncing (Git) %q to (OCI) %q: %w", action.GitRemote, action.Repo, err)
	}
	rootUI.Infof("Commit Manifest digest: %s", commitManDesc.Digest)

	if err := toOCI.Cleanup(); err != nil {
		return fmt.Errorf("cleaning up toOCI: %w", err)
	}

	return nil
}
