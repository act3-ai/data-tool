package mirror

import (
	"context"
	"fmt"
	"sync/atomic"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/act3-ai/data-tool/internal/mirror"
	dtreg "github.com/act3-ai/data-tool/internal/registry"
	"github.com/act3-ai/data-tool/internal/ui"
	"github.com/act3-ai/go-common/pkg/logger"
)

// Gather represents the mirror gather action.
type Gather struct {
	*Action

	// IndexFallback is set when the target registry does not support index-of-index behavior.
	// It will push the nested index to the target repository and add its reference to the annotations of the main gather index.
	IndexFallback bool

	// ExtraAnnotations defines the user-created annotations to add to the index of the gather repository.
	ExtraAnnotations map[string]string

	// Platforms defines the platform(s) for the images to be gathered. (Default behavior is to gather all available platforms.)
	Platforms []string
}

// Run executes the actual gather operation.
func (action *Gather) Run(ctx context.Context, sourceFile string, dest string) error {
	log := logger.FromContext(ctx)
	cfg := action.Config.Get(ctx)

	rootUI := ui.FromContextOrNoop(ctx)

	// initialize extra annotations if it is not set
	if action.ExtraAnnotations == nil {
		action.ExtraAnnotations = make(map[string]string)
	}

	destTarget, err := action.Config.GraphTarget(ctx, dest)
	if err != nil {
		return err
	}

	destRef, err := dtreg.ParseEndpointOrDefault(action.Config, dest)
	if err != nil {
		return fmt.Errorf("resolving destination reference with endpoint resolution: %w", err)
	}

	// create the gather opts
	opts := mirror.GatherOptions{
		Platforms:      action.Platforms,
		ConcurrentHTTP: cfg.ConcurrentHTTP,
		DestStorage:    destTarget,
		Log:            log,
		RootUI:         rootUI,
		SourceFile:     sourceFile,
		Dest:           dest,
		Annotations:    action.ExtraAnnotations,
		IndexFallback:  action.IndexFallback,
		DestReference:  destRef,
		Recursive:      action.Recursive,
		Targeter:       action.Config,
	}

	// run the gather function
	idxDesc, err := mirror.Gather(ctx, action.Version(), opts)
	if err != nil {
		return fmt.Errorf("gathering artifacts: %w", err)
	}

	err = destTarget.Tag(ctx, idxDesc, destRef.Reference)
	if err != nil {
		return fmt.Errorf("tagging gather index manifest: %w", err)
	}
	referenceWithDigest := opts.DestReference
	referenceWithDigest.Reference = idxDesc.Digest.String()
	opts.RootUI.Infof("Gather index: %s", referenceWithDigest.String())
	opts.RootUI.Infof("Pushed index to destination: %s", destRef)

	return nil
}

// WorkTracker is an object for tracking the number of blobs and bytes actually pushed.
type WorkTracker struct {
	blobs       atomic.Int64
	transferred atomic.Int64
}

// Add adds the digest and blob to the work tracker count.
func (wt *WorkTracker) Add(desc ocispec.Descriptor) {
	wt.blobs.Add(1)
	wt.transferred.Add(desc.Size)
}
