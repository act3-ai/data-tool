package bottle

import (
	"context"
	"fmt"
	"io"

	"github.com/act3-ai/data-tool/internal/bottle"
	"github.com/act3-ai/go-common/pkg/logger"
)

// Status represents the bottle status action.
type Status struct {
	*Action

	// Show file paths within subdirectories for changed files
	Details bool
}

// Run runs the bottle status action.
func (action *Status) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	cfg := action.Config.Get(ctx)

	log.InfoContext(ctx, "bottle status command activated")
	btlPath, err := bottle.FindBottleRootDir(action.Dir)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "loading bottle information from specified path", "path", btlPath)
	btl, err := bottle.LoadBottle(btlPath, bottle.WithCachePath(cfg.CachePath))
	if err != nil {
		return err
	}

	statusStr, _, err := bottle.InspectBottleFiles(ctx, btl, bottle.Options{WantDetails: action.Details})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(out, statusStr)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "bottle status command completed")

	return nil
}
