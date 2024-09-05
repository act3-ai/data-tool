// Package bottle defines bottle command and subcommand actions.
package bottle

import (
	"context"
	"fmt"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// Action represents a general bottle action.
type Action struct {
	*actions.DataTool

	Dir string // Bottle directory
}

// CompressionLevelOptions defines options for bottle part Compression Levels.
type CompressionLevelOptions struct {
	Level string // change the compression level of compressed bottle parts
}

// saveMetaChanges performs saves the changes made to the metadata and saves them on disk. It also logs its progress
// while doing so. Errors are return if fails in the process.
func saveMetaChanges(ctx context.Context, btl *bottle.Bottle) error {
	log := logger.FromContext(ctx)

	// btlMetadataSaveOpts is a SaveOptions constant used when only bottle metadata is updated in a bottle
	var btlMetadataSaveOpts = bottle.SaveOptions{
		NoArchive: true,
		NoDigest:  true,
		NoCommit:  true,
	}

	// save changes to bottle's metadata
	err := bottle.SaveUpdatesToSet(ctx, btl, btlMetadataSaveOpts)

	if err != nil {
		return fmt.Errorf("failed while saving bottle at %s: %w", btl.GetPath(), err)
	}

	log.InfoContext(ctx, "command complete")
	return nil
}

func (action *Action) prepare(ctx context.Context) (*v1alpha1.Configuration, *bottle.Bottle, error) {
	cfg := action.Config.Get(ctx)

	var err error
	action.Dir, err = bottle.FindBottleRootDir(action.Dir)
	if err != nil {
		return cfg, nil, fmt.Errorf("failed to find root bottle directory starting from %s: %w", action.Dir, err)
	}

	btl, err := LoadAndUpgradeBottle(ctx, cfg, action.Dir)
	if err != nil {
		return cfg, nil, fmt.Errorf("failed to load bottle at %s: %w", action.Dir, err)
	}

	return cfg, btl, nil
}

// LoadAndUpgradeBottle loads a bottle from the provided path, and checks if
// it needs to be updated using the migration system.
func LoadAndUpgradeBottle(ctx context.Context, cfg *v1alpha1.Configuration, path string) (*bottle.Bottle, error) {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "loading bottle information from specified path", "path", path)

	btl, err := bottle.LoadBottle(path,
		bottle.WithCachePath(cfg.CachePath),
		bottle.WithBlobInfoCache(cfg.CachePath),
	)
	if err != nil {
		return nil, err
	}

	// TODO it would be nice to only walk the directory tree once
	// not in InspectBottleFiles and when we read labels.
	_, _, err = bottle.InspectBottleFiles(ctx, btl, bottle.Options{Visitor: bottle.PrepareUpdatedParts(ctx, btl)})
	if err != nil {
		return nil, fmt.Errorf("failed while checking for updated bottle parts: %w", err)
	}

	// we do this in bottle.LoadBottle() but that is too early (we do not have all our parts at that time)
	if err := btl.LoadLocalLabels(); err != nil {
		return nil, err
	}

	return btl, nil
}
