package bottle

import (
	"context"
	"fmt"
	"io"

	"gitlab.com/act3-ai/asce/data/tool/internal/actions/internal/format"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// ArtifactList represents the bottle artifact list action.
type ArtifactList struct {
	*Action
}

// Run runs the bottle artifact list action.
func (action *ArtifactList) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	artifacts := btl.Definition.PublicArtifacts
	if len(artifacts) == 0 {
		log.InfoContext(ctx, "bottle has no public artifact to show", "path", action.Dir)
		return nil
	}

	// Print artifact table list
	t := format.NewTable()
	t.AddRow("NAME", "MEDIA TYPE", "PATH")

	for _, art := range artifacts {
		t.AddRow(art.Name, art.MediaType, art.Path)
	}

	_, err = fmt.Fprintln(out, t.String())
	if err != nil {
		return err
	}

	return nil
}
