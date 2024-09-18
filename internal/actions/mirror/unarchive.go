package mirror

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/cache"
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

	// Reference is an optional reference to tag the image in disk storage. If not set, "latest" will be used.
	Reference string
}

// Run runs the mirror unarchive action.
func (action *Unarchive) Run(ctx context.Context, sourceFile, mappingSpec string) error {
	log := logger.FromContext(ctx)
	cfg := action.Config.Get(ctx)

	// must enable with predecessors before deserialize, if we want scatter to utilize it later
	storage, err := oci.NewStorage(cfg.CachePath)
	if err != nil {
		return fmt.Errorf("initializing cache storage: %w", err)
	}
	gstorage := cache.NewPredecessorCacher(storage)

	rootUI := ui.FromContextOrNoop(ctx)

	// create the deserialize options
	deserializeOptions := mirror.DeserializeOptions{
		DestStorage: gstorage,
		DestTargetReference: registry.Reference{
			Reference: action.Reference,
		},
		SourceFile: sourceFile,
		BufferSize: action.BufferSize,
		DryRun:     action.DryRun,
		RootUI:     rootUI,
		Strict:     action.Strict,
		Log:        log,
	}

	// run deserialize
	idxDesc, err := mirror.Deserialize(ctx, deserializeOptions)
	if err != nil {
		return fmt.Errorf("deserializing: %w", err)
	}

	// create the scatter options
	scatterOptions := mirror.ScatterOptions{
		SubsetFile: action.SubsetFile,
		Source:     gstorage,
		SourceDesc: idxDesc,
		SourceReference: registry.Reference{
			Reference: action.Reference,
		},
		MappingSpec:    mappingSpec,
		Selectors:      action.Selectors,
		ConcurrentHTTP: cfg.ConcurrentHTTP,
		RootUI:         rootUI,
		DryRun:         action.DryRun,
		Recursive:      action.Recursive,
		Targeter:       action.Config,
	}

	// run scatter
	return mirror.Scatter(ctx, scatterOptions)
}
