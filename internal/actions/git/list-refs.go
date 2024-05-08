package git

import (
	"context"
	"fmt"
	"path/filepath"

	"gitlab.com/act3-ai/asce/data/tool/internal/git"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// ListRefs represents the git list-refs action.
type ListRefs struct {
	*Action

	Repo string
}

// Run performs the list-refs operation.
func (action *ListRefs) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	log.InfoContext(ctx, "Configuring OCI repository")
	repo, err := action.Config.ConfigureRepository(ctx, action.Repo)
	if err != nil {
		return fmt.Errorf("creating repository reference: %w", err)
	}

	fromOCI, err := git.NewFromOCI(ctx, repo, repo.Reference.Reference, "", git.SyncOptions{}, &cmd.Options{})
	if err != nil {
		return fmt.Errorf("prepparing to run from-oci action: %w", err)
	}
	defer fromOCI.Cleanup() //nolint

	log.InfoContext(ctx, "fetching base manifest and config")
	manDesc, err := fromOCI.FetchBaseManifestConfig(ctx)
	if err != nil {
		return fmt.Errorf("fetching base manifest and config: %w", err)
	}

	rootUI.Infof("Digest of %s: %s", action.Repo, manDesc.Digest)
	rootUI.Infof("References:")

	tags, err := fromOCI.GetTagRefs()
	if err != nil {
		return fmt.Errorf("getting tag references: %w", err)
	}

	heads, err := fromOCI.GetHeadRefs()
	if err != nil {
		return fmt.Errorf("getting head references: %w", err)
	}

	for tag, refInfo := range tags {
		rootUI.Infof("%s %s", refInfo.Commit, filepath.Join(cmd.TagRefPrefix, tag))
	}

	for head, refInfo := range heads {
		rootUI.Infof("%s %s", refInfo.Commit, filepath.Join(cmd.HeadRefPrefix, head))
	}

	if err := fromOCI.Cleanup(); err != nil {
		return fmt.Errorf("cleaning up fromOCI: %w", err)
	}

	return nil
}
