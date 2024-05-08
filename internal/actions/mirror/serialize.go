package mirror

import (
	"context"

	"git.act3-ace.com/ace/data/tool/internal/mirror"
)

// Serialize represents the mirror serialize action.
type Serialize struct {
	*Action

	Checkpoint          string                    // path to save the checkpoint file
	ExistingCheckpoints []mirror.ResumeFromLedger // a slice of existing checkpoint files in the case of multiple failures
}

// Run runs the mirror serialize action.
func (action *Serialize) Run(ctx context.Context, ref string, destFile string, existingImages []string, n, bs, hwm int) error {

	repo, err := action.Config.ConfigureRepository(ctx, ref)
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
		RepoFunc:            action.Config.ConfigureRepository,
		SourceRepo:          repo,
		SourceReference:     repo.Reference.ReferenceOrDefault(),
	}

	return mirror.Serialize(ctx, destFile, action.Checkpoint, action.DataTool.Version(), opts)
}
