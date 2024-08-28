package mirror

import (
	"context"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
)

// Clone represents the mirror clone action.
type Clone struct {
	*Action

	// Display repository manifest destinations, but do not push
	Check bool

	// Scatter images filtered by labels in annotations
	Selectors []string

	// Platforms defines the platform(s) for the images to be gathered. (Default behavior is to gather all available platforms.)
	Platforms []string
}

// Run runs the mirror clone action.
func (action *Clone) Run(ctx context.Context, sourceFile, mappingSpec string) error {
	log := logger.FromContext(ctx)
	cfg := action.Config.Get(ctx)

	rootUI := ui.FromContextOrNoop(ctx)

	// create clone opts
	opts := mirror.CloneOptions{
		MappingSpec:    mappingSpec,
		Selectors:      action.Selectors,
		ConcurrentHTTP: cfg.ConcurrentHTTP,
		Platforms:      action.Platforms,
		Log:            log,
		SourceFile:     sourceFile,
		RootUI:         rootUI,
		Targeter:       action.Config,
		Recursive:      action.Recursive,
		DryRun:         action.Check,
	}

	// run mirror clone
	return mirror.Clone(ctx, opts)
}
