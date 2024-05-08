package mirror

import (
	"context"

	"git.act3-ace.com/ace/data/tool/internal/mirror"
	"git.act3-ace.com/ace/data/tool/internal/ui"
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

	src, err := action.Config.ConfigureRepository(ctx, sourceRepo)
	if err != nil {
		return err
	}

	// create the scatter options
	opts := mirror.ScatterOptions{
		SubsetFile:     action.SourceFile,
		Src:            src,
		SrcString:      sourceRepo,
		SrcReference:   src.Reference,
		MappingSpec:    mappingSpec,
		Selectors:      action.Selectors,
		ConcurrentHTTP: cfg.ConcurrentHTTP,
		RootUI:         rootUI,
		DryRun:         action.Check,
		Recursive:      action.Recursive,
		RepoFunc:       action.Config.ConfigureRepository,
	}

	// run mirror scatter
	return mirror.Scatter(ctx, opts)
}
