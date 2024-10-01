package mirror

import (
	"context"
	"fmt"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/cache"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"

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

	// WithManifestJSON specifies whether or not to write out a manifest.json file, similar to 'docker image save'.
	WithManifestJSON bool

	// ExtraAnnotations defines the user-created annotations to add to the index of the gather repository.
	ExtraAnnotations map[string]string

	// Platforms defines the platform(s) for the images to be gathered. (Default behavior is to gather all available platforms.)
	Platforms []string

	// Compression defines the compression type (zstd and gzip supported)
	Compression string
	// Reference is an optional reference to tag the image in disk storage. If not set, "latest" will be used.
	Reference string
}

// Run executes the actual archive operation.
func (action *Archive) Run(ctx context.Context, sourceFile, destFile string, existingImages []string, n, bs, hwm int) error {

	log := logger.FromContext(ctx)
	cfg := action.Config.Get(ctx)

	// wrapping the cache Storage with in-memory predecessors upgrades it to a
	// GraphStorage. This aids in satisfying the requirement that our gather here
	// does not push/gather to a remote but is limited to pulling the blobs locally.
	// We cannot rely on the multiple remote sources in serialize, as it expects
	// to serialize from a single source.
	storage, err := oci.NewStorage(cfg.CachePath)
	if err != nil {
		return fmt.Errorf("initializing cache storage: %w", err)
	}
	gstorage := cache.NewPredecessorCacher(storage)

	rootUI := ui.FromContextOrNoop(ctx)

	// create the gather opts
	gatherOpts := mirror.GatherOptions{
		Platforms:      action.Platforms,
		ConcurrentHTTP: cfg.ConcurrentHTTP,
		DestStorage:    gstorage,
		Log:            log,
		RootUI:         rootUI,
		SourceFile:     sourceFile,
		Dest:           destFile,
		Annotations:    action.ExtraAnnotations,
		IndexFallback:  action.IndexFallback,
		DestReference:  registry.Reference{Reference: action.Reference},
		Recursive:      action.Recursive,
		Targeter:       action.Config,
	}

	// run the gather function
	idxDesc, err := mirror.Gather(ctx, action.DataTool.Version(), gatherOpts)
	if err != nil {
		return err
	}

	// create serialize options
	options := mirror.SerializeOptions{
		BufferOpts:          mirror.BlockBufOptions{Buffer: n, BlockSize: bs, HighWaterMark: hwm},
		ExistingCheckpoints: action.ExistingCheckpoints,
		ExistingImages:      existingImages,
		Recursive:           action.Recursive,
		RepoFunc:            action.Config.Repository,
		Compression:         action.Compression,
		SourceStorage:       gstorage,
		SourceReference:     action.Reference,
		SourceDesc:          idxDesc,
		WithManifestJSON:    action.WithManifestJSON,
	}
	// serialize it
	return mirror.Serialize(ctx, destFile, action.Checkpoint, action.DataTool.Version(), options)
}
