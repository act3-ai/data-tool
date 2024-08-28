package mirror

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2/registry"

	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
)

// Scatter represents the mirror scatter action.
type Scatter struct {
	*Action

	Check      bool     // Display repository manifest destinations, but do not push
	SourceFile string   // The optional sources.list can be passed to scatter a subset of the images from the source repository (i.e., not all of the images in the source repository).
	Selectors  []string // Scatter images filtered by labels in annotations
}

// Run runs the mirror scatter action.
func (action *Scatter) Run(ctx context.Context, sourceRepo, mappingSpec string) error {
	// log := logger.FromContext(ctx)

	cfg := action.Config.Get(ctx)

	rootUI := ui.FromContextOrNoop(ctx)

	gtarget, err := action.Config.ReadOnlyGraphTarget(ctx, sourceRepo)
	if err != nil {
		return err
	}

	srcDesc, err := gtarget.Resolve(ctx, sourceRepo)
	if err != nil {
		return fmt.Errorf("resolving source reference: %w", err)
	}

	srcRef, err := registry.ParseReference(sourceRepo)
	if err != nil {
		return fmt.Errorf("parsing destination reference: %w", err)
	}

	// create the scatter options
	opts := mirror.ScatterOptions{
		SubsetFile:      action.SourceFile,
		Source:          gtarget,
		SourceDesc:      srcDesc,
		SourceReference: srcRef,
		MappingSpec:     mappingSpec,
		Selectors:       action.Selectors,
		ConcurrentHTTP:  cfg.ConcurrentHTTP,
		RootUI:          rootUI,
		DryRun:          action.Check,
		Recursive:       action.Recursive,
		Targeter:        action.Config,
	}

	// run mirror scatter
	return mirror.Scatter(ctx, opts)
}
