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

// FromOCI represents the git rebuild action.
type FromOCI struct {
	*Action

	Repo      string
	GitRemote string
	Force     bool
}

// Run performs the FromOCI operation.
func (action *FromOCI) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	log.InfoContext(ctx, "Configuring repository", "repo", action.Repo)
	repo, err := action.Config.Repository(ctx, action.Repo)
	if err != nil {
		return fmt.Errorf("creating repository reference: %w", err)
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "FromOCI-*") // removed by fromOCI.Cleanup
	if err != nil {
		return fmt.Errorf("creating temporary directory for intermediate git repository: %w", err)
	}

	syncOpts := git.SyncOptions{
		DTVersion: action.Version(),
		TmpDir:    tmpDir,
		CacheDir:  action.CacheDir,
	}

	cmdOpts := cmd.Options{
		GitOptions: cmd.GitOptions{
			Force:      action.Force,
			AltGitExec: action.AltGitExec,
		},
		LFSOptions: &cmd.LFSOptions{
			WithLFS:    action.LFS,
			AltLFSExec: action.AltGitLFSExec,
			ServerURL:  action.LFSServerURL,
		},
	}

	fromOCI, err := git.NewFromOCI(ctx, repo, repo.Reference.Reference, action.GitRemote, syncOpts, &cmdOpts)
	if err != nil {
		return fmt.Errorf("prepparing to run from-oci action: %w", err)
	}
	defer fromOCI.Cleanup() //nolint

	log.InfoContext(ctx, "Starting FromOCI action")
	updatedRefs, err := fromOCI.Run(ctx)
	if err != nil {
		return fmt.Errorf("rebuilding (OCI) %q to (Git) %q: %w", action.Repo, action.GitRemote, err)
	}

	rootUI.Infof("Git repository update complete. The following references have been updated:\n")
	for _, entry := range updatedRefs {
		rootUI.Infof("%s", entry)
	}

	if err := fromOCI.Cleanup(); err != nil {
		return fmt.Errorf("cleaning up fromOCI: %w", err)
	}

	return nil
}
