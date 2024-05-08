package bottle

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/opencontainers/go-digest"

	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// Commit represents the bottle commit action.
type Commit struct {
	*Action

	Write       WriteBottleOptions
	Compression CompressionLevelOptions
	NoDeprecate bool // Don't deprecate existing bottle
}

// Run runs the bottle commit action.
func (action *Commit) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "bottle commit command activated")

	cfg, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	return commit(ctx, cfg, btl, action.Write.BottleIDFile, action.NoDeprecate)
}

func commit(ctx context.Context, cfg *v1alpha1.Configuration, btl *bottle.Bottle, bottleIDFile string, noDeprecate bool) error {
	log := logger.FromContext(ctx)

	rootUI := ui.FromContextOrNoop(ctx)

	// for each public artifact in the bottle path ensure we have the digest calculated
	if err := verifyArtifacts(btl, rootUI); err != nil {
		return err
	}

	// Before we commit, check if there is a bottleID in the .dt directory,
	// This would mean that the bottle has already been committed / pulled / pushed
	// and ace-dt now infers this a deprecating a bottle.
	// Get bottleID from bottleID file if it exists
	bottlePath := btl.GetPath()
	defaultBottleIDPath := filepath.Join(bottlePath, ".dt", "bottleid")
	var oldDigest digest.Digest

	// check the bottlePath for a bottleID file, if it exists, then ReadBottleIDFile
	if _, err := os.Stat(defaultBottleIDPath); err == nil {
		oldDigest, err = bottle.ReadBottleIDFile(bottlePath)
		if err != nil {
			return err
		}
	}

	// Access global flag compressionLevel
	level := cfg.CompressionLevel

	log.InfoContext(ctx, "Saving updated bottle")
	if err := SaveUpdatesToSet(ctx, btl, SaveOptions{CompressLevel: level}); err != nil {
		return err
	}

	log.InfoContext(ctx, "Validating bottle data")
	if err := btl.Definition.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("failed to validate bottle before push: %w", err)
	}

	// if noDeprecate flag isn't set, then we deprecate the previous bottleID
	// If the bottleID changed after saving the updates to the set,
	// then we deprecate the previous bottleID (from default bottleID file read earlier)
	if !noDeprecate && oldDigest.String() != "" && oldDigest != btl.GetBottleID() {
		// save changes to bottle, adding the deprecated to the entry.yaml (bottle config)
		if err := deprecate(log, rootUI, btl, oldDigest); err != nil {
			return err
		}
		if err := btl.ConstructManifest(); err != nil { // reconstruct the updated information
			return fmt.Errorf("resetting bottle configuration: %w", err)
		}
	}

	log.InfoContext(ctx, "bottle commit command complete")

	return bottle.SaveExtraBottleInfo(ctx, btl, bottleIDFile)
}

// deprecate deprecates the previous bottleID, then saves the new bottle configuration.
//
//nolint:sloglint
func deprecate(log *slog.Logger, u *ui.Task, btl *bottle.Bottle, dgst digest.Digest) error {
	log.Info("BottleID changed, deprecating previous bottleID")
	btl.DeprecateBottleID(dgst)

	if err := btl.Save(); err != nil {
		return err
	}
	u.Infof("Deprecating previous bottle %s", dgst)
	return nil
}

// verifyArtifacts ensures that all public artifacts have a digest calculated.
func verifyArtifacts(btl *bottle.Bottle, u *ui.Task) error {
	for i, art := range btl.Definition.PublicArtifacts {
		if art.Digest == "" {
			if _, err := os.Stat(btl.NativePath(art.Path)); errors.Is(err, fs.ErrNotExist) {
				u.Infof("PublicArtifact %q not found.\nPlease update bottle metadata to reference an existing file.", art.Path)
				continue
			}
			if err := btl.CalculatePublicArtifactDigest(i); err != nil {
				return err
			}
		}
	}
	return nil
}
