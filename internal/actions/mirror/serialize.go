package mirror

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2/registry"

	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
)

// Serialize represents the mirror serialize action.
type Serialize struct {
	*Action

	Checkpoint          string                    // path to save the checkpoint file
	ExistingCheckpoints []mirror.ResumeFromLedger // a slice of existing checkpoint files in the case of multiple failures
	Compression         string                    // compression type (zstd and gzip supported)

	// WithManifestJSON specifies whether or not to write out a manifest.json file, similar to 'docker image save'.
	WithManifestJSON bool
}

// Run runs the mirror serialize action.
func (action *Serialize) Run(ctx context.Context, ref string, destFile string, existingImages []string, n, bs, hwm int) error {

	gt, err := action.Config.GraphTarget(ctx, ref)
	if err != nil {
		return err
	}

	rr, err := registry.ParseReference(ref)
	if err != nil {
		return fmt.Errorf("parsing registry reference: %w", err)
	}

	sourceRef := rr.ReferenceOrDefault()
	sourceDesc, err := gt.Resolve(ctx, sourceRef)
	if err != nil {
		return fmt.Errorf("getting remote descriptor for %s: %w", sourceRef, err)
	}

	// create the Serialize Options
	opts := mirror.SerializeOptions{
		BufferOpts: mirror.BlockBufOptions{
			Buffer:        n,
			BlockSize:     bs,
			HighWaterMark: hwm,
		},
		ExistingCheckpoints: action.ExistingCheckpoints,
		ExistingImages:      existingImages,
		Recursive:           action.Recursive,
		RepoFunc:            action.Config.Repository,
		SourceStorage:       gt,
		SourceReference:     sourceRef,
		SourceDesc:          sourceDesc,
		Compression:         action.Compression,
		WithManifestJSON:    action.WithManifestJSON,
	}

	return mirror.Serialize(ctx, destFile, action.Checkpoint, action.DataTool.Version(), opts)
}
