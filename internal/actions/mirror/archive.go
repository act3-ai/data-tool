package mirror

import (
	"context"
	"fmt"

	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
)

// Archive represents the mirror clone action.
type Archive struct {
	*Action

	// Only archive the images filtered by labels in annotations
	Selectors []string

	// Checkpoint is the path to save the checkpoint file
	Checkpoint string

	// ExistingCheckpoints is a slice of existing checkpoint files in the case of multiple failures
	ExistingCheckpoints []mirror.ResumeFromLedger

	// IndexFallback is set when the target registry does not support index-of-index behavior.
	// It will push the nested index to the target repository and add its reference to the annotations of the main gather index.
	IndexFallback bool

	// ExtraAnnotations defines the user-created annotations to add to the index of the gather repository.
	ExtraAnnotations map[string]string

	// Platforms defines the platform(s) for the images to be gathered. (Default behavior is to gather all available platforms.)
	Platforms []string
}

// Run executes the actual archive operation.
func (action *Archive) Run(ctx context.Context, sourceFile, destFile, reference string, existingImages []string, n, bs, hwm int) error {
	log := logger.FromContext(ctx)
	cfg := action.Config.Get(ctx)

	store, err := oci.NewWithContext(ctx, cfg.CachePath)
	if err != nil {
		return fmt.Errorf("error creating oci store: %w", err)
	}

	rootUI := ui.FromContextOrNoop(ctx)

	// create the gather opts
	gatherOpts := mirror.GatherOptions{
		Platforms:      action.Platforms,
		ConcurrentHTTP: cfg.ConcurrentHTTP,
		DestTarget:     store,
		Log:            log,
		RootUI:         rootUI,
		SourceFile:     sourceFile,
		Dest:           destFile,
		Annotations:    action.ExtraAnnotations,
		IndexFallback:  action.IndexFallback,
		DestReference:  registry.Reference{Reference: reference},
		Recursive:      action.Recursive,
		RepoFunc:       action.Config.Repository,
	}

	// run the gather function
	if err := mirror.Gather(ctx, action.DataTool.Version(), gatherOpts); err != nil {
		return err
	}

	// create serialize options
	options := mirror.SerializeOptions{
		BufferOpts:          mirror.BlockBufOptions{Buffer: n, BlockSize: bs, HighWaterMark: hwm},
		ExistingCheckpoints: action.ExistingCheckpoints,
		ExistingImages:      existingImages,
		Recursive:           action.Recursive,
		RepoFunc:            action.Config.Repository,
		SourceRepo:          store,
		SourceReference:     reference,
	}
	// serialize it
	return mirror.Serialize(ctx, destFile, action.Checkpoint, action.DataTool.Version(), options)
}
