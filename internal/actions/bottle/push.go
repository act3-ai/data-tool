package bottle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	sigcustom "git.act3-ace.com/ace/data/tool/internal/sign"

	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/data/telemetry/pkg/types"
	"git.act3-ace.com/ace/data/tool/internal/actions"
	"git.act3-ace.com/ace/data/tool/internal/bottle"
	"git.act3-ace.com/ace/data/tool/internal/bottle/status"
	"git.act3-ace.com/ace/data/tool/internal/ref"
	"git.act3-ace.com/ace/data/tool/internal/storage"
	tbottle "git.act3-ace.com/ace/data/tool/internal/transfer/bottle"
	"git.act3-ace.com/ace/data/tool/internal/ui"
	tbtl "git.act3-ace.com/ace/data/tool/pkg/transfer/bottle"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// Push represents the bottle push action.
type Push struct {
	*Action

	Write       WriteBottleOptions
	Telemetry   actions.TelemetryOptions
	Compression CompressionLevelOptions

	NoOverwrite bool // Only push data if if doesn't already exist
	NoDeprecate bool // Don't deprecate existing bottle

	Ref string
}

// Run runs the bottle push action.
func (action *Push) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	log.InfoContext(ctx, "bottle push command activated")

	cfg, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	// first we must commit, this saves everything: manifest, config, archived parts, etc.
	log.InfoContext(ctx, "committing bottle")
	if err := commit(ctx, cfg, btl, action.Write.BottleIDFile, action.NoDeprecate); err != nil {
		return err
	}

	if action.NoOverwrite {
		log.InfoContext(ctx, "checking for existing bottles")
		parsedref := ref.RepoFromString(action.Ref)
		repo, err := action.Config.ConfigureRepository(ctx, parsedref.String())
		if err != nil {
			return fmt.Errorf("creating repository reference: %w", err)
		}

		_, err = repo.Resolve(ctx, repo.Reference.Reference)
		switch {
		case errors.Is(err, errdef.ErrNotFound):
			// continue
		case err != nil:
			// unwanted error
			return fmt.Errorf("checking if reference exists: %w", err)
		default:
			// reference already exists
			return fmt.Errorf("bottle reference %s already exists. Please choose another repository or tag before pushing. %w",
				action.Ref, err)
		}

	}

	transferCfg := tbtl.NewTransferConfig(ctx, action.Ref, action.Dir, action.Config)

	log.InfoContext(ctx, "pushing bottle with signatures")
	if err := tbottle.PushBottle(ctx, btl, *transferCfg, tbottle.WithSignatures()); err != nil {
		return fmt.Errorf("pushing bottle and signatures: %w", err)
	}

	r, err := ref.FromString(action.Ref)
	if err != nil {
		return fmt.Errorf("invalid bottle reference %s: %w", action.Ref, err)
	}

	// Handle telemetry
	telem := bottle.NewTelemetryAdapter(cfg.Telemetry, cfg.TelemetryUserName)

	// send telemetry
	telemUrls, err := telem.SendTelemetry(logger.NewContext(ctx, logger.V(log, 1)), btl, r, types.EventPush)
	if err != nil {
		return err
	}

	summary, err := sigcustom.NewSummaryFromBottle(ctx, btl)
	if err != nil {
		return fmt.Errorf("generating signature detail message: %w", err)
	}
	if summary != nil {
		err = telem.SendSignatures(logger.NewContext(ctx, logger.V(log, 1)), summary)
		if err != nil {
			return err
		}
	}

	rootUI.Info(formatBottleURLs(telemUrls))

	rootUI.Infof("Bottle push complete.  BottleID: %s\n", btl.GetBottleID())

	return nil
}

// SaveOptions is a structure for supplying options to the SaveUpdatesToSet function. By default, all options
// are "on", eg, the options to disable functions are all false.
type SaveOptions struct {
	NoArchive     bool
	NoDigest      bool
	NoCommit      bool
	CompressLevel string
}

// SaveUpdatesToSet performs archival, digest, and cache commission to bottle components, and saves bottle metadata.
func SaveUpdatesToSet(ctx context.Context, btl *bottle.Bottle, options SaveOptions) error {
	var tmpFileMap sync.Map

	log := logger.FromContext(ctx)

	if !options.NoArchive {
		log.InfoContext(ctx, "Checking if files need to be archived")
		if err := archiveParts(ctx, btl, options.CompressLevel, &tmpFileMap); err != nil {
			return err
		}
	}
	if !options.NoDigest {
		log.InfoContext(ctx, "Checking if files need to be digested")
		if err := digestParts(ctx, btl); err != nil {
			return err
		}
	}
	if !options.NoCommit {
		log.InfoContext(ctx, "Committing new files to cache")
		if err := commitParts(ctx, btl, &tmpFileMap); err != nil {
			return err
		}

		// build the latest manifest for the bottle based on any updated information generated above. We don't need the
		// manifest handler here, but note that it is saved within the bottle.
		err := btl.ConstructManifest()
		if err != nil {
			return err
		}
	}

	log.InfoContext(ctx, "Saving updated information to bottle")
	return btl.Save()
}

// prepareUpdatedParts performs bottle part processing as a delegate when scanning for changed parts.  The bottle
// information is updated with the changed data, preserving existing data where possible.   Mostly, this involves
// removing file entries, resetting file entries (removing size/digest to trigger recalc), and adding file entries.
func prepareUpdatedParts(ctx context.Context, btl *bottle.Bottle) status.Visitor {
	fsys := os.DirFS(btl.GetPath())

	// TODO why does this function not return an error, return an error
	return func(info storage.PartInfo, status bottle.PartStatus) (bool, error) {
		name := info.GetName()
		log := logger.FromContext(ctx).With("name", name)

		switch status {
		case bottle.StatusDeleted:
			log.InfoContext(ctx, "Bottle part removed")
			btl.RemovePartMetadata(name)
		case bottle.StatusChanged:
			log.InfoContext(ctx, "Changed part flagged for reprocessing")
			// TODO why do we not reset the content size?
			/// Seems like we need to be setting content size somewhere else (where we update everything else), not preserving it here.
			btl.UpdatePartMetadata(name,
				info.GetContentSize(), "",
				nil, // preserve part labels
				info.GetLayerSize(), "",
				"",
				&time.Time{},
			)
		case bottle.StatusNew:
			log.InfoContext(ctx, "New part flagged for processing")
			fullPath := btl.NativePath(name)
			fInfo, err := os.Stat(fullPath)
			if err != nil {
				log.InfoContext(ctx, "file discovered during push, but not able to stat")
				return false, err
			}

			if strings.HasSuffix(name, "/") {
				if err := addDirToBottle(ctx, fsys, btl, name, fInfo); err != nil {
					return false, fmt.Errorf("unable to add directory to bottle at %s: %w", fullPath, err)
				}
			} else {
				if err := addFileToBottle(ctx, btl, name, fInfo); err != nil {
					return false, fmt.Errorf("unable to add file to bottle %s: %w", fullPath, err)
				}
			}
		case bottle.StatusCached, bottle.StatusDigestMatch, bottle.StatusExists, bottle.StatusVirtual:
			return false, nil
		}
		return false, nil
	}
}
