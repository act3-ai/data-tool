package actions

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gitlab.com/act3-ai/asce/data/tool/pkg/cache"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// Prune represents the prune action.
type Prune struct {
	*DataTool

	Max int64 // Maximum size to keep in the cache, in MiB
}

// Run runs the prune action.
func (action *Prune) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "util prune command activated")

	cfg := action.Config.Get(ctx)

	max, valid := cfg.CachePruneMax.AsInt64()
	if !valid {
		return fmt.Errorf("invalid config value cachePruneMax of %s", cfg.CachePath)
	}
	// check the flag
	if action.Max >= 0 {
		max = action.Max
	}
	// max is specified in megabytes
	max = max * 1024 * 1024
	log.InfoContext(ctx, "Max size selected", "size", max)

	cacheManager := cache.NewBottleCachePruner(cfg.CachePath)

	if err := cacheManager.Prune(max); err != nil {
		return fmt.Errorf("error pruning cache: %w", err)
	}
	_, err := fmt.Fprintf(out, "Cache pruned to maximum size: %d MiB\n", max/(1024*1024))
	if err != nil {
		return err
	}

	// Normally this will be empty, but early termination could cause items to linger here
	// we will recreate this directory another time, can just delete everything in it
	if err := os.RemoveAll(filepath.Join(cfg.CachePath, "scratch")); err != nil {
		return fmt.Errorf("failed to remove scratch directory: %w", err)
	}

	log.InfoContext(ctx, "util prune command completed")
	return nil
}
