package mirror

import (
	"context"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
)

// Deserialize represents the mirror serialize action.
type Deserialize struct {
	*Action

	// Strict ensures that the order of files is correct in the tar stream.
	Strict bool

	// DryRun indicates that data is not to be sent to target registry (data will be discarded instead).
	DryRun bool

	// BufferSize defines the number of bytes to use for the the buffer for reading from the archive (tape)
	BufferSize int
}

// Run runs the mirror deserialize action.
func (action *Deserialize) Run(ctx context.Context, sourceFile string, dest string) error {
	rootUI := ui.FromContextOrNoop(ctx)

	log := logger.FromContext(ctx)

	repo, err := action.Config.Repository(ctx, dest)
	if err != nil {
		return err
	}
	// create deserialize options
	opts := mirror.DeserializeOptions{
		DestTarget:          repo,
		DestTargetReference: repo.Reference,
		SourceFile:          sourceFile,
		BufferSize:          action.BufferSize,
		DryRun:              action.DryRun,
		RootUI:              rootUI,
		Strict:              action.Strict,
		Log:                 log,
	}

	// run mirror deserialize
	return mirror.Deserialize(ctx, opts)
}
