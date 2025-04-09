package mirror

import (
	"context"
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/act3-ai/data-tool/internal/mirror"
	"github.com/act3-ai/data-tool/internal/ui"
	"github.com/act3-ai/go-common/pkg/logger"
)

// Deserialize represents the mirror deserialize action.
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

	gt, err := action.Config.GraphTarget(ctx, dest)
	if err != nil {
		return err
	}

	// parse with endpoint resolution
	destRef, err := action.Config.ParseEndpointReference(dest)
	if err != nil {
		return fmt.Errorf("parsing destination reference: %w", err)
	}

	// create deserialize options
	opts := mirror.DeserializeOptions{
		DestStorage:         gt,
		DestTargetReference: destRef,
		SourceFile:          sourceFile,
		BufferSize:          action.BufferSize,
		DryRun:              action.DryRun,
		RootUI:              rootUI,
		Strict:              action.Strict,
		Log:                 log,
	}

	// run mirror deserialize
	idxDesc, err := mirror.Deserialize(ctx, opts)
	if err != nil {
		return fmt.Errorf("deserializing: %w", err)
	}

	// Tag it based on the input from the user
	tag := opts.DestTargetReference.ReferenceOrDefault()
	opts.RootUI.Infof("Tagging root node as %q", tag)
	if err := gt.Tag(ctx, idxDesc, tag); err != nil {
		return fmt.Errorf("tagging the %q file: %w", ocispec.ImageIndexFile, err)
	}

	return nil
}
