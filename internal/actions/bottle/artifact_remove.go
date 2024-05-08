package bottle

import (
	"context"
	"io"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// ArtifactRemove represents the bottle artifact remove action.
type ArtifactRemove struct {
	*Action
}

// Run runs the bottle artifact remove action.
func (action *ArtifactRemove) Run(ctx context.Context, artifactPath string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "artifact remove command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	err = btl.RemoveArtifact(artifactPath)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "removed artifact from public artifact list of the bottle", "path", artifactPath, "bottlePath", btl.GetPath())
	log.InfoContext(ctx, "Saving bottle with specified public artifact removed")

	return saveMetaChanges(ctx, btl)
}
