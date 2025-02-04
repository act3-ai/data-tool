package bottle

import (
	"context"
	"fmt"
	"io"

	"git.act3-ace.com/ace/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/data/tool/internal/util"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// ArtifactSet represents the bottle artifact set action.
type ArtifactSet struct {
	*Action

	MediaType string // Artifact's media type
}

// Run runs the bottle artifact set action.
func (action *ArtifactSet) Run(ctx context.Context, artName, artPath string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "artifact add command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	artDigest, err := util.DigestFile(artPath)
	if err != nil {
		return fmt.Errorf("artifact digest: %w", err)
	}

	// check if media-type was specified, otherwise try and deduce media type
	if action.MediaType == "" {
		action.MediaType = mediatype.DetermineType(artPath)
	}

	if err := btl.AddArtifact(artName, artPath, action.MediaType, artDigest); err != nil {
		return err
	}

	return saveMetaChanges(ctx, btl)
}
