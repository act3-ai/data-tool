package bottle

import (
	"context"
	"fmt"
	"io"
	"os"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// Describe represents the bottle describe action.
type Describe struct {
	*Action

	File string // File with description text to add to a bottle
}

// Run runs the bottle describe action.
func (action *Describe) Run(ctx context.Context, description string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "describe command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	// Make sure a description was passed
	if description == "" && action.File == "" {
		return fmt.Errorf("missing arguments or flags for command")
	}

	// Set the description via the options
	if description == "" && action.File != "" {
		data, err := os.ReadFile(action.File)
		if err != nil {
			return fmt.Errorf("failed to read description file %s: %w", action.File, err)
		}
		description = string(data)
	}

	// Add the annotation to the bottle.
	btl.SetDescription(description)

	log.InfoContext(ctx, "describe command completed")

	return saveMetaChanges(ctx, btl)
}
