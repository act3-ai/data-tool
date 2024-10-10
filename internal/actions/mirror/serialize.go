package mirror

import (
	"context"
	"fmt"

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

	// parse with endpoint resolution
	rr, err := action.Config.ParseEndpointReference(ref)
	if err != nil {
		return fmt.Errorf("parsing registry reference: %w", err)
	}

	// ensure we pass the full reference in the case gt is an endpointResolver
	sourceRef := rr.ReferenceOrDefault()
	rr.Reference = sourceRef
	sourceDesc, err := gt.Resolve(ctx, rr.String())
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
