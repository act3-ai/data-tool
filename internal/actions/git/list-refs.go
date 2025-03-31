package git

import (
	"context"
	"fmt"
	"path"

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
	repo, err := action.Config.Repository(ctx, action.Repo)
	if err != nil {
		return fmt.Errorf("creating repository reference: %w", err)
	}

	desc, err := repo.Resolve(ctx, repo.Reference.Reference)
	if err != nil {
		return fmt.Errorf("resolving base manifest descriptor: %w", err)
	}

	fromOCI, err := git.NewFromOCI(ctx, repo, desc, "", git.SyncOptions{}, &cmd.Options{})
	if err != nil {
		return fmt.Errorf("prepparing to run from-oci action: %w", err)
	}
	defer fromOCI.Cleanup() //nolint

	log.InfoContext(ctx, "fetching base manifest and config")
	err = fromOCI.FetchBaseManifestConfig(ctx)
	if err != nil {
		return fmt.Errorf("fetching base manifest and config: %w", err)
	}

	rootUI.Infof("Digest of %s: %s", action.Repo, desc.Digest)
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
		rootUI.Infof("%s %s", refInfo.Commit, path.Join(cmd.TagRefPrefix, tag)) // references don't use OS-specific path separators
	}

	for head, refInfo := range heads {
		rootUI.Infof("%s %s", refInfo.Commit, path.Join(cmd.HeadRefPrefix, head)) // references don't use OS-specific path separators
	}

	if err := fromOCI.Cleanup(); err != nil {
		return fmt.Errorf("cleaning up fromOCI: %w", err)
	}

	return nil
}
