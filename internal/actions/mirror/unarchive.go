package mirror

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
)

// Unarchive represents the mirror unarchive action.
type Unarchive struct {
	*Action

	// Strict ensures that the order of files is correct in the tar stream.
	Strict bool

	// DryRun indicates that data is not to be sent to target registry (data will be discarded instead).
	DryRun bool

	// BufferSize defines the number of bytes to use for the the buffer for reading from the archive (tape)
	BufferSize int

	// An optional sources.list can be passed to scatter a subset of the images from the source repository (i.e., not all of the images in the source repository).
	SubsetFile string

	// Scatter images filtered by labels in annotations
	Selectors []string
}

// Run runs the mirror unarchive action.
func (action *Unarchive) Run(ctx context.Context, sourceFile, mappingSpec, reference string) error {
	log := logger.FromContext(ctx)
	cfg := action.Config.Get(ctx)

	// create the oci.Store
	store, err := oci.NewWithContext(ctx, cfg.CachePath)
	if err != nil {
		return fmt.Errorf("error creating oci store: %w", err)
	}

	rootUI := ui.FromContextOrNoop(ctx)

	// create the deserialize options
	deserializeOptions := mirror.DeserializeOptions{
		DestTarget: store,
		DestTargetReference: registry.Reference{
			Reference: reference,
		},
		SourceFile: sourceFile,
		BufferSize: action.BufferSize,
		DryRun:     action.DryRun,
		RootUI:     rootUI,
		Strict:     action.Strict,
		Log:        log,
	}

	// run deserialize
	if err := mirror.Deserialize(ctx, deserializeOptions); err != nil {
		return err
	}

	// create the scatter options
	scatterOptions := mirror.ScatterOptions{
		SubsetFile: action.SubsetFile,
		Src:        store,
		SrcString:  cfg.CachePath,
		SrcReference: registry.Reference{
			Reference: reference,
		},
		MappingSpec:    mappingSpec,
		Selectors:      action.Selectors,
		ConcurrentHTTP: cfg.ConcurrentHTTP,
		RootUI:         rootUI,
		DryRun:         action.DryRun,
		Recursive:      action.Recursive,
		RepoFunc:       action.Config.Repository,
	}

	// run scatter
	return mirror.Scatter(ctx, scatterOptions)
}
