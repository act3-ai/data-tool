package bottle

import (
	"context"
	"fmt"
	"io"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle/status"
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

	statusStr, _, err := status.InspectBottleFiles(ctx, btl, status.Options{WantDetails: action.Details})
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
