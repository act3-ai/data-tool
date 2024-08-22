package mirror

import (
	"context"

	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
)

// Serialize represents the mirror serialize action.
type Serialize struct {
	*Action

	Checkpoint          string                    // path to save the checkpoint file
	ExistingCheckpoints []mirror.ResumeFromLedger // a slice of existing checkpoint files in the case of multiple failures
	Compression         string                    // compression type (zstd and gzip supported)
}

// Run runs the mirror serialize action.
func (action *Serialize) Run(ctx context.Context, ref string, destFile string, existingImages []string, n, bs, hwm int) error {
	repo, err := action.Config.Repository(ctx, ref)
	if err != nil {
		return err
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
		SourceRepo:          repo,
		SourceReference:     repo.Reference.ReferenceOrDefault(),
		Compression:         action.Compression,
	}

	return mirror.Serialize(ctx, destFile, action.Checkpoint, action.DataTool.Version(), opts)
}
